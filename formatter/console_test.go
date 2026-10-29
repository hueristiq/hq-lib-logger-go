package formatter

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestConsoleDefaults(t *testing.T) {
	t.Parallel()

	log := &Log{
		Time:    time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC),
		Level:   levels.LevelInfo,
		Message: "hello",
		Fields:  []Field{String("b", "2"), String("a", "1")},
	}

	data, err := NewConsole(nil).Format(log)

	require.NoError(t, err)
	assert.Equal(t, "2026-10-28T12:00:00Z [INF] hello a=1 b=2", string(data))
}

func TestConsoleZeroValue(t *testing.T) {
	t.Parallel()

	data, err := (&Console{}).Format(&Log{
		Level:   levels.LevelWarn,
		Message: "plain",
		Fields:  []Field{String("z", "1"), String("a", "2")},
	})

	require.NoError(t, err)
	assert.Equal(t, "plain a=2 z=1", string(data))
}

func TestConsoleNameCallerAndError(t *testing.T) {
	t.Parallel()

	formatter := NewConsole(&ConsoleConfig{Label: true})

	data, err := formatter.Format(&Log{
		Level:   levels.LevelError,
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

	data, err := formatter.Format(&Log{
		Level:   levels.LevelInfo,
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
		Colorizer: ColorizerFunc(func(text string, _ levels.Level) string { return "<" + text + ">" }),
	})

	data, err := formatter.Format(&Log{Level: levels.LevelInfo, Message: "hello"})

	require.NoError(t, err)
	assert.Equal(t, "[<INF>] hello", string(data))
}

func TestConsoleInvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := NewConsole(nil).Format(&Log{Level: levels.Level(99)})

	require.ErrorIs(t, err, levels.ErrUnknownLevel)
}
