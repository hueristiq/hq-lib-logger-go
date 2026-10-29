package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestNoOpColorizer(t *testing.T) {
	t.Parallel()

	var colorizer Colorizer = NoOpColorizer{}

	assert.Equal(t, "INF", colorizer.Colorize("INF", levels.LevelInfo))
}

func TestColorizerFunc(t *testing.T) {
	t.Parallel()

	var nilColorizer Colorizer = ColorizerFunc(nil)

	assert.Equal(t, "text", nilColorizer.Colorize("text", levels.LevelInfo))

	colorizer := ColorizerFunc(func(text string, _ levels.Level) string {
		return "<" + text + ">"
	})

	assert.Equal(t, "<text>", colorizer.Colorize("text", levels.LevelInfo))
}
