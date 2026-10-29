package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestNoOpColorizer(t *testing.T) {
	t.Parallel()

	var colorizer Colorizer = NoOpColorizer{}

	assert.Equal(t, "INF", colorizer.Colorize("INF", hqgologgerlevels.LevelInfo))
}

func TestColorizerFunc(t *testing.T) {
	t.Parallel()

	var nilColorizer Colorizer = ColorizerFunc(nil)

	assert.Equal(t, "text", nilColorizer.Colorize("text", hqgologgerlevels.LevelInfo))

	colorizer := ColorizerFunc(func(text string, _ hqgologgerlevels.Level) string {
		return "<" + text + ">"
	})

	assert.Equal(t, "<text>", colorizer.Colorize("text", hqgologgerlevels.LevelInfo))
}
