//nolint:testpackage // tests exercise unexported behavior.
package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

func TestNoOpColorizer(t *testing.T) {
	t.Parallel()

	var colorizer Colorizer = NoOpColorizer{}

	assert.Equal(t, "INF", colorizer.Colorize("INF", hqgologgerlevel.LevelInfo))
}

func TestColorizerFunc(t *testing.T) {
	t.Parallel()

	var nilColorizer Colorizer = ColorizerFunc(nil)

	assert.Equal(t, "text", nilColorizer.Colorize("text", hqgologgerlevel.LevelInfo))

	colorizer := ColorizerFunc(func(text string, _ hqgologgerlevel.Level) string {
		return "<" + text + ">"
	})

	assert.Equal(t, "<text>", colorizer.Colorize("text", hqgologgerlevel.LevelInfo))
}
