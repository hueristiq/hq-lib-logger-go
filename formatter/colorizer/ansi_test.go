package colorizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestANSIColorizerColoredLevels(t *testing.T) {
	t.Parallel()

	c := NewANSIColorizer()
	require.NotNil(t, c)

	tests := []struct {
		name   string
		level  hqgologgerlevels.Level
		params string
	}{
		{"fatal", hqgologgerlevels.LevelFatal, "1;91"},
		{"error", hqgologgerlevels.LevelError, "1;91"},
		{"info", hqgologgerlevels.LevelInfo, "1;94"},
		{"warn", hqgologgerlevels.LevelWarn, "1;93"},
		{"debug", hqgologgerlevels.LevelDebug, "1;95"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := c.Colorize("INF", tt.level)

			assert.Equal(t, "\x1b["+tt.params+"mINF\x1b[0m", got)
		})
	}
}

func TestANSIColorizerSilentAndInvalidLevelsPassThrough(t *testing.T) {
	t.Parallel()

	c := NewANSIColorizer()

	for _, level := range []hqgologgerlevels.Level{
		hqgologgerlevels.LevelSilent,
		hqgologgerlevels.Level(-1),
	} {
		assert.Equal(t, "INF", c.Colorize("INF", level))
	}
}

func TestANSIColorizerImplementsColorizer(t *testing.T) {
	t.Parallel()

	var colorizer interface {
		Colorize(text string, level hqgologgerlevels.Level) (colorized string)
	} = NewANSIColorizer()

	assert.NotNil(t, colorizer)
}
