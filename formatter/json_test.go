package formatter

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestJSONDefaults(t *testing.T) {
	t.Parallel()

	data, err := NewJSON(nil).Format(Log{
		Time:    time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC),
		Level:   levels.LevelInfo,
		Message: "hello",
		Fields:  []Field{String("b", "2"), String("a", "1")},
	})

	require.NoError(t, err)
	assert.True(t, json.Valid(data))

	var decoded map[string]any

	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "2026-10-28T12:00:00Z", decoded["time"])
	assert.Equal(t, "info", decoded["level"])
	assert.Equal(t, "hello", decoded["message"])
	assert.Equal(t, "1", decoded["a"])
	assert.Equal(t, "2", decoded["b"])
}

func TestJSONCustomKeysAndNestedFields(t *testing.T) {
	t.Parallel()

	formatter := NewJSON(&JSONConfig{
		LevelKey:   "lvl",
		MessageKey: "msg",
		FieldsKey:  "fields",
		ErrorKey:   "err",
	})

	data, err := formatter.Format(Log{
		Level:   levels.LevelWarn,
		Message: "warned",
		Fields: []Field{
			String("k", "v"),
			Err("err", errors.New("boom")),
		},
	})

	require.NoError(t, err)

	var decoded map[string]any

	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "warn", decoded["lvl"])
	assert.Equal(t, "warned", decoded["msg"])
	assert.Equal(t, "boom", decoded["err"])
	assert.Equal(t, map[string]any{"k": "v"}, decoded["fields"])
}

func TestJSONZeroConfigOmitsOptionalParts(t *testing.T) {
	t.Parallel()

	data, err := NewJSON(&JSONConfig{}).Format(Log{
		Time:    time.Now(),
		Level:   levels.LevelInfo,
		Message: "hello",
		Name:    "api",
		Caller:  "main.go:1",
		Fields:  []Field{Err("error", errors.New("boom"))},
	})

	require.NoError(t, err)

	var decoded map[string]any

	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.NotContains(t, decoded, "time")
	assert.NotContains(t, decoded, "logger")
	assert.NotContains(t, decoded, "caller")
	assert.Equal(t, "boom", decoded["error"])
}

func TestJSONEscapesTimestampFormat(t *testing.T) {
	t.Parallel()

	formatter := NewJSON(&JSONConfig{
		Timestamp:       true,
		TimestampFormat: `2006"01"02`,
	})

	data, err := formatter.Format(Log{
		Time:    time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC),
		Level:   levels.LevelInfo,
		Message: "hello",
	})

	require.NoError(t, err)
	assert.True(t, json.Valid(data))

	var decoded map[string]any

	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, `2026"10"28`, decoded["time"])
}

func TestJSONPreservesFieldOrder(t *testing.T) {
	t.Parallel()

	formatter := NewJSON(&JSONConfig{PreserveFieldOrder: true})

	data, err := formatter.Format(Log{
		Level:   levels.LevelInfo,
		Message: "m",
		Fields:  []Field{String("z", "1"), String("a", "2")},
	})

	require.NoError(t, err)
	assert.Contains(t, string(data), `"z":"1","a":"2"`)
}

func TestJSONInvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := NewJSON(nil).Format(Log{Level: levels.Level(99)})

	require.ErrorIs(t, err, levels.ErrUnknownLevel)
}

func TestJSONAppendVariants(t *testing.T) {
	t.Parallel()

	cfg := &JSONConfig{
		Timestamp:  false,
		LevelKey:   "level",
		MessageKey: "message",
		NameKey:    "logger",
		CallerKey:  "caller",
		ErrorKey:   "error",
		FieldsKey:  "fields",
	}
	jsonFormatter := NewJSON(cfg)

	base := Log{Level: levels.LevelInfo, Message: "m"}

	withError := base
	withError.Fields = []Field{Err("error", errors.New("boom"))}

	data, err := jsonFormatter.Append(nil, withError)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"error":"boom"`)

	nested := base
	nested.Fields = []Field{String("k", "v")}

	data, err = jsonFormatter.Append(nil, nested)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"fields":{"k":"v"}`)

	withErrAny := base
	withErrAny.Fields = []Field{Any("k", errors.New("e"))}

	data, err = jsonFormatter.Append(nil, withErrAny)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"k":"e"`)

	namedAndLocated := base
	namedAndLocated.Name = "api"
	namedAndLocated.Caller = "main.go:1"

	data, err = jsonFormatter.Append(nil, namedAndLocated)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"logger":"api"`)
	assert.Contains(t, string(data), `"caller":"main.go:1"`)

	withUnmarshalable := base
	withUnmarshalable.Fields = []Field{Any("k", make(chan int))}

	data, err = jsonFormatter.Append(nil, withUnmarshalable)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"k":`)
}

func TestJSONAppendIndent(t *testing.T) {
	t.Parallel()

	jsonFormatter := NewJSON(&JSONConfig{
		Timestamp:  false,
		LevelKey:   "level",
		MessageKey: "message",
		Indent:     "  ",
	})

	data, err := jsonFormatter.Append(nil, Log{Level: levels.LevelInfo, Message: "m"})

	require.NoError(t, err)
	assert.Contains(t, string(data), "\n  \"level\": \"info\"")
}

func TestJSONLayoutSafe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		layout string
		want   bool
	}{
		{"rfc3339", time.RFC3339, true},
		{"plain layout", "Jan 2 15:04:05", true},
		{"quote", `2006"01"02`, false},
		{"backslash", `2006\01`, false},
		{"control", "2006\x0101", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, jsonLayoutSafe(tt.layout))
		})
	}
}

func TestJSONAppendManyFields(t *testing.T) {
	t.Parallel()

	formatter := NewJSON(nil)

	fields := make([]Field, 0, 22)
	fields = append(fields, String("dup", "a"), String("dup", "b"))

	for i := range 20 {
		fields = append(fields, Int("k"+strconv.Itoa(i), i))
	}

	data, err := formatter.Append(nil, Log{Level: levels.LevelInfo, Message: "m", Fields: fields})

	require.NoError(t, err)
	assert.Contains(t, string(data), `"k19":19`)
	assert.Contains(t, string(data), `"dup":"b"`)
}
