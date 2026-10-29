package formatter

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestLogfmtDefaults(t *testing.T) {
	t.Parallel()

	data, err := NewLogfmt(nil).Format(Log{
		Level:   levels.LevelInfo,
		Message: "hello world",
		Fields:  []Field{String("b", "2"), String("a", "1")},
	})

	require.NoError(t, err)
	assert.Equal(t, `level=info msg="hello world" a=1 b=2`, string(data))
}

func TestLogfmtCustomKeysAndError(t *testing.T) {
	t.Parallel()

	formatter := NewLogfmt(&LogfmtConfig{
		LevelKey:   "lvl",
		MessageKey: "msg",
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
	assert.Equal(t, `lvl=warn msg=warned err=boom k=v`, string(data))
}

func TestLogfmtQuotesUnsafeValues(t *testing.T) {
	t.Parallel()

	data, err := NewLogfmt(&LogfmtConfig{}).Format(Log{
		Level:   levels.LevelInfo,
		Message: "line\nbreak",
		Fields:  []Field{String("path", `a=b`), String("empty", "")},
	})

	require.NoError(t, err)
	assert.Contains(t, string(data), `msg="line\nbreak"`)
	assert.Contains(t, string(data), `empty=`)
	assert.Contains(t, string(data), `path="a=b"`)
}

func TestLogfmtInvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := NewLogfmt(nil).Format(Log{Level: levels.Level(99)})

	require.ErrorIs(t, err, levels.ErrUnknownLevel)
}

func TestLogfmtEdges(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `"un safe"`, string(appendLogfmtKey(nil, "un safe")))
	assert.Equal(t, "key=", string(appendLogfmtValue([]byte("key="), "")))
	assert.Equal(t, `key="a=b"`, string(appendLogfmtValue([]byte("key="), "a=b")))
	assert.Equal(t, `key="a b"`, string(appendLogfmtValue([]byte("key="), "a b")))
	assert.Equal(t, `key="a\xffb"`, string(appendLogfmtValue([]byte("key="), "a\xffb")))

	errField := Err("k", errors.New("boom"))
	assert.Equal(t, "boom", string(appendLogfmtFieldValue(nil, &errField)))

	anyField := Any("k", "x")
	assert.Equal(t, "x", string(appendLogfmtFieldValue(nil, &anyField)))

	intField := Int("k", 42)
	assert.Equal(t, "42", string(appendLogfmtFieldValue(nil, &intField)))

	assert.False(t, logfmtSafe("a b"))
	assert.False(t, logfmtSafe("a=b"))
	assert.False(t, logfmtSafe("a\"b"))
	assert.False(t, logfmtSafe("a\\b"))
	assert.False(t, logfmtSafe("a\nb"))
	assert.False(t, logfmtSafe("a\xffb"))
	assert.True(t, logfmtSafe("abc"))
}

func TestLogfmtAppendFullRecord(t *testing.T) {
	t.Parallel()

	cfg := &LogfmtConfig{
		Timestamp:       true,
		TimestampFormat: time.RFC3339,
		TimestampKey:    "time",
		LevelKey:        "level",
		MessageKey:      "msg",
		NameKey:         "logger",
		CallerKey:       "caller",
		ErrorKey:        "error",
	}
	logfmtFormatter := NewLogfmt(cfg)

	record := Log{
		Time:    time.Unix(0, 0).UTC(),
		Level:   levels.LevelInfo,
		Message: "m",
		Name:    "api",
		Caller:  "main.go:1",
	}

	data, err := logfmtFormatter.Append(nil, record)

	require.NoError(t, err)

	output := string(data)
	assert.Contains(t, output, `time=1970-01-01T00:00:00Z`)
	assert.Contains(t, output, `level=info`)
	assert.Contains(t, output, `msg=m`)
	assert.Contains(t, output, `logger=api`)
	assert.Contains(t, output, `caller=main.go:1`)
}

func TestLogfmtAppendUnsafeTimestampLayout(t *testing.T) {
	t.Parallel()

	formatter := NewLogfmt(&LogfmtConfig{
		Timestamp:       true,
		TimestampFormat: "Jan 2 2006",
		TimestampKey:    "time",
		LevelKey:        "level",
		MessageKey:      "msg",
	})

	data, err := formatter.Append(nil, Log{
		Time:    time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC),
		Level:   levels.LevelInfo,
		Message: "m",
	})

	require.NoError(t, err)
	assert.Contains(t, string(data), `time="Oct 29 2026"`)
}

func TestLogfmtAppendMixedOrderFields(t *testing.T) {
	t.Parallel()

	formatter := NewLogfmt(nil)

	data, err := formatter.Append(nil, Log{
		Level:   levels.LevelInfo,
		Message: "m",
		Fields:  []Field{String("z", "1"), String("a", "2"), String("a", "4"), String("m", "3")},
	})

	require.NoError(t, err)

	output := string(data)
	assert.Contains(t, output, "a=2")
	assert.Contains(t, output, "a=4")
	assert.Contains(t, output, "m=3")
	assert.Contains(t, output, "z=1")
	assert.Less(t, strings.Index(output, "a=2"), strings.Index(output, "m=3"))
	assert.Less(t, strings.Index(output, "m=3"), strings.Index(output, "z=1"))
}

func TestLogfmtAppendSingleField(t *testing.T) {
	t.Parallel()

	formatter := NewLogfmt(nil)

	data, err := formatter.Append(nil, Log{
		Level:   levels.LevelInfo,
		Message: "m",
		Fields:  []Field{String("only", "1")},
	})

	require.NoError(t, err)
	assert.Contains(t, string(data), "only=1")
}
