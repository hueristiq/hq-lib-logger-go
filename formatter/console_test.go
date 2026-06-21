package formatter

import (
	"errors"
	"strings"
	"testing"
	"time"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type upperColorizer struct{}

func (upperColorizer) Colorize(text string, _ hqgologgerlevels.Level) string {
	return strings.ToUpper(text)
}

func plainFormatter() *Console {
	return NewConsoleFormatter(&ConsoleFormatterConfiguration{
		IncludeTimestamp: false,
		IncludeLabel:     true,
		Colorize:         false,
		Colorizer:        NewNoOpColorizer(),
	})
}

func TestNewConsoleFormatterNilUsesDefaults(t *testing.T) {
	t.Parallel()

	f := NewConsoleFormatter(nil)
	require.NotNil(t, f)

	data, err := f.Format(&Log{
		Timestamp: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC),
		Level:     hqgologgerlevels.LevelInfo,
		Message:   "hello",
		Metadata:  map[string]any{"label": "INF"},
	})

	require.NoError(t, err)

	assert.Equal(t, "2025-01-02T03:04:05Z [INF] hello", string(data))
}

func TestFormatBasicMessage(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:   hqgologgerlevels.LevelInfo,
		Message: "just a message",
	})

	require.NoError(t, err)

	assert.Equal(t, "just a message", string(data))
}

func TestFormatWithLabel(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:    hqgologgerlevels.LevelWarn,
		Message:  "careful",
		Metadata: map[string]any{"label": "WRN"},
	})

	require.NoError(t, err)

	assert.Equal(t, "[WRN] careful", string(data))
}

func TestFormatEmptyLabelOmitted(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:    hqgologgerlevels.LevelInfo,
		Message:  "no label",
		Metadata: map[string]any{"label": ""},
	})

	require.NoError(t, err)

	assert.Equal(t, "no label", string(data))
	assert.NotContains(t, string(data), "[]")
}

func TestFormatLabelDisabledByConfig(t *testing.T) {
	t.Parallel()

	f := NewConsoleFormatter(&ConsoleFormatterConfiguration{
		IncludeLabel: false,
		Colorizer:    NewNoOpColorizer(),
	})

	data, err := f.Format(&Log{
		Level:    hqgologgerlevels.LevelInfo,
		Message:  "msg",
		Metadata: map[string]any{"label": "INF", "k": "v"},
	})

	require.NoError(t, err)

	assert.Equal(t, "msg k=v", string(data))
}

func TestFormatColorizeInvokesColorizer(t *testing.T) {
	t.Parallel()

	f := NewConsoleFormatter(&ConsoleFormatterConfiguration{
		IncludeLabel: true,
		Colorize:     true,
		Colorizer:    upperColorizer{},
	})

	data, err := f.Format(&Log{
		Level:    hqgologgerlevels.LevelInfo,
		Message:  "m",
		Metadata: map[string]any{"label": "inf"},
	})

	require.NoError(t, err)

	assert.Equal(t, "[INF] m", string(data))
}

func TestFormatTimestampIncludedAndFormatted(t *testing.T) {
	t.Parallel()

	f := NewConsoleFormatter(&ConsoleFormatterConfiguration{
		IncludeTimestamp: true,
		TimestampFormat:  "2006-01-02 15:04:05",
		Colorizer:        NewNoOpColorizer(),
	})

	data, err := f.Format(&Log{
		Timestamp: time.Date(2025, 8, 8, 13, 45, 0, 0, time.UTC),
		Level:     hqgologgerlevels.LevelInfo,
		Message:   "m",
	})

	require.NoError(t, err)

	assert.Equal(t, "2025-08-08 13:45:00 m", string(data))
}

func TestFormatZeroTimestampOmitted(t *testing.T) {
	t.Parallel()

	f := NewConsoleFormatter(&ConsoleFormatterConfiguration{
		IncludeTimestamp: true,
		TimestampFormat:  time.RFC3339,
		Colorizer:        NewNoOpColorizer(),
	})

	data, err := f.Format(&Log{
		Level:   hqgologgerlevels.LevelInfo,
		Message: "m",
	})

	require.NoError(t, err)

	assert.Equal(t, "m", string(data))
}

func TestFormatMetadataDeterministicOrder(t *testing.T) {
	t.Parallel()

	log := &Log{
		Level:   hqgologgerlevels.LevelInfo,
		Message: "m",
		Metadata: map[string]any{
			"label": "INF",
			"zeta":  "1",
			"alpha": "2",
			"mike":  "3",
		},
	}

	want := "[INF] m alpha=2 mike=3 zeta=1"

	for range 50 {
		data, err := plainFormatter().Format(log)
		require.NoError(t, err)
		assert.Equal(t, want, string(data))
	}
}

func TestFormatMetadataSkipsEmptyKeyAndNilValue(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:   hqgologgerlevels.LevelInfo,
		Message: "m",
		Metadata: map[string]any{
			"":     "ignored",
			"nilv": nil,
			"keep": "yes",
		},
	})

	require.NoError(t, err)

	assert.Equal(t, "m keep=yes", string(data))
}

func TestFormatNonStringMetadataValue(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:    hqgologgerlevels.LevelInfo,
		Message:  "m",
		Metadata: map[string]any{"count": 42, "ratio": 1.5},
	})

	require.NoError(t, err)

	assert.Equal(t, "m count=42 ratio=1.5", string(data))
}

func TestFormatErrorRenderedOnce(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:   hqgologgerlevels.LevelError,
		Message: "boom",
		Metadata: map[string]any{
			"label": "ERR",
			"error": errors.New("connection timeout"),
		},
	})

	require.NoError(t, err)

	out := string(data)

	assert.NotContains(t, out, "error=")
	assert.Equal(t, 1, strings.Count(out, "connection timeout"))
	assert.Equal(t, "[ERR] boom\n\nconnection timeout", out)
}

func TestFormatErrorAlongsideOtherMetadata(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:   hqgologgerlevels.LevelError,
		Message: "boom",
		Metadata: map[string]any{
			"label": "ERR",
			"user":  "alex",
			"error": errors.New("nope"),
		},
	})

	require.NoError(t, err)

	assert.Equal(t, "[ERR] boom user=alex\n\nnope", string(data))
}

func TestFormatNonErrorErrorKey(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:    hqgologgerlevels.LevelError,
		Message:  "boom",
		Metadata: map[string]any{"label": "ERR", "error": "stringy"},
	})

	require.NoError(t, err)

	assert.Equal(t, "[ERR] boom\n\nstringy", string(data))
}

func TestFormatInvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := plainFormatter().Format(&Log{
		Level:   hqgologgerlevels.Level(99),
		Message: "x",
	})

	require.Error(t, err)

	assert.NotContains(t, err.Error(), "%!w")
	assert.NotContains(t, err.Error(), "<nil>")
	assert.Contains(t, err.Error(), "99")
}

func TestFormatTrimsTrailingNewline(t *testing.T) {
	t.Parallel()

	data, err := plainFormatter().Format(&Log{
		Level:   hqgologgerlevels.LevelInfo,
		Message: "trailing\n",
	})

	require.NoError(t, err)

	assert.Equal(t, "trailing", string(data))
}

func TestDefaultConsoleConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConsoleConfig()

	require.NotNil(t, cfg)

	assert.True(t, cfg.IncludeTimestamp)
	assert.Equal(t, time.RFC3339, cfg.TimestampFormat)
	assert.True(t, cfg.IncludeLabel)
	assert.True(t, cfg.Colorize)
	assert.NotNil(t, cfg.Colorizer)
	assert.False(t, cfg.PrettyPrint)
}

func TestConsoleImplementsFormatter(t *testing.T) {
	t.Parallel()

	var _ Formatter = NewConsoleFormatter(nil)
}
