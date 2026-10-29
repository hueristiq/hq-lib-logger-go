package formatter

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestConsoleDefaults(t *testing.T) {
	t.Parallel()

	log := &Log{
		Time:    time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC),
		Level:   hqgologgerlevels.LevelInfo,
		Message: "hello",
		Fields:  []Field{String("b", "2"), String("a", "1")},
	}

	data, err := NewConsole(nil).Format(*log)

	require.NoError(t, err)
	assert.Equal(t, "2026-10-28T12:00:00Z [INF] hello a=1 b=2", string(data))
}

func TestConsoleZeroValue(t *testing.T) {
	t.Parallel()

	data, err := (&Console{}).Format(Log{
		Level:   hqgologgerlevels.LevelWarn,
		Message: "plain",
		Fields:  []Field{String("z", "1"), String("a", "2")},
	})

	require.NoError(t, err)
	assert.Equal(t, "plain a=2 z=1", string(data))
}

func TestConsoleNameCallerAndError(t *testing.T) {
	t.Parallel()

	formatter := NewConsole(&ConsoleConfig{Label: true})

	data, err := formatter.Format(Log{
		Level:   hqgologgerlevels.LevelError,
		Message: "failed",
		Name:    "api",
		Caller:  "main.go:12",
		Fields: []Field{
			Err(ErrorKey, errors.New("boom")),
			String("request_id", "42"),
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "[ERR] [api] failed caller=main.go:12 request_id=42\n\nboom", string(data))
}

func TestConsolePreservesFieldOrder(t *testing.T) {
	t.Parallel()

	formatter := NewConsole(&ConsoleConfig{PreserveFieldOrder: true})

	data, err := formatter.Format(Log{
		Level:   hqgologgerlevels.LevelInfo,
		Message: "ordered",
		Fields:  []Field{String("z", "1"), String("a", "2")},
	})

	require.NoError(t, err)
	assert.Equal(t, "ordered z=1 a=2", string(data))
}

func TestConsoleColorizesLabel(t *testing.T) {
	t.Parallel()

	formatter := NewConsole(&ConsoleConfig{
		Label:     true,
		Colorize:  true,
		Colorizer: ColorizerFunc(func(text string, _ hqgologgerlevels.Level) string { return "<" + text + ">" }),
	})

	data, err := formatter.Format(Log{Level: hqgologgerlevels.LevelInfo, Message: "hello"})

	require.NoError(t, err)
	assert.Equal(t, "[<INF>] hello", string(data))
}

func TestConsoleInvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := NewConsole(nil).Format(Log{Level: hqgologgerlevels.Level(99)})

	require.ErrorIs(t, err, hqgologgerlevels.ErrUnknownLevel)
}

func TestConsoleColorizeDefaultsToNoOp(t *testing.T) {
	t.Parallel()

	console := NewConsole(&ConsoleConfig{Colorize: true})

	assert.IsType(t, NoOpColorizer{}, console.cfg.Colorizer)
}

func TestConsoleAppendEdges(t *testing.T) {
	t.Parallel()

	cfg := &ConsoleConfig{Timestamp: false, Label: true}
	console := NewConsole(cfg)

	tests := []struct {
		name string
		log  Log
		want string
	}{
		{
			"non-string label field",
			Log{Level: hqgologgerlevels.LevelInfo, Message: "m", Fields: []Field{Int(LabelKey, 5)}},
			"m",
		},
		{
			"custom string label field",
			Log{Level: hqgologgerlevels.LevelInfo, Message: "m", Fields: []Field{String(LabelKey, "CUSTOM")}},
			"[CUSTOM] m",
		},
		{
			"caller key field skipped",
			Log{Level: hqgologgerlevels.LevelInfo, Message: "m", Caller: "main.go:1", Fields: []Field{String(CallerKey, "override")}},
			"[INF] m caller=main.go:1",
		},
		{
			"empty message with caller",
			Log{Level: hqgologgerlevels.LevelInfo, Message: "", Caller: "main.go:1"},
			"[INF] caller=main.go:1",
		},
		{
			"nil error field",
			Log{Level: hqgologgerlevels.LevelInfo, Message: "m", Fields: []Field{Err(ErrorKey, nil)}},
			"[INF] m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := console.Append(nil, tt.log)

			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}
}

func TestConsoleAppendDuplicateKeys(t *testing.T) {
	t.Parallel()

	console := NewConsole(&ConsoleConfig{Timestamp: false, Label: true})

	data, err := console.Append(nil, Log{
		Level:   hqgologgerlevels.LevelInfo,
		Message: "m",
		Fields:  []Field{String("dup", "a"), String("dup", "b")},
	})

	require.NoError(t, err)
	assert.Contains(t, string(data), "dup=b")
}
