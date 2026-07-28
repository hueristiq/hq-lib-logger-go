package formatter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestNoOpColorizerReturnsInputUnchanged(t *testing.T) {
	t.Parallel()

	c := hqgologgerformatter.NewNoOpColorizer()

	require.NotNil(t, c)

	for _, level := range []hqgologgerlevels.Level{
		hqgologgerlevels.LevelFatal,
		hqgologgerlevels.LevelSilent,
		hqgologgerlevels.LevelError,
		hqgologgerlevels.LevelInfo,
		hqgologgerlevels.LevelWarn,
		hqgologgerlevels.LevelDebug,
	} {
		assert.Equal(t, "INF", c.Colorize("INF", level))
	}
}

func TestNoOpColorizerImplementsColorizer(t *testing.T) {
	t.Parallel()

	var _ hqgologgerformatter.Colorizer = hqgologgerformatter.NewNoOpColorizer()
}
