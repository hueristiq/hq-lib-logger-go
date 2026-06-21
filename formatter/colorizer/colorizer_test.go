package colorizer

import (
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	color.NoColor = false

	os.Exit(m.Run())
}

type colorizerFactory struct {
	name string
	new  func() hqgologgerformatter.Colorizer
}

func factories() []colorizerFactory {
	return []colorizerFactory{
		{"fatih", func() hqgologgerformatter.Colorizer { return NewFatihColorizer() }},
		{"aurora", func() hqgologgerformatter.Colorizer { return NewAuroraColorizer() }},
	}
}

func TestColorizerImplementsInterface(t *testing.T) {
	t.Parallel()

	var (
		_ hqgologgerformatter.Colorizer = NewFatihColorizer()
		_ hqgologgerformatter.Colorizer = NewAuroraColorizer()
	)
}

func TestColorizeWrapsColoredLevels(t *testing.T) {
	t.Parallel()

	colored := []hqgologgerlevels.Level{
		hqgologgerlevels.LevelFatal,
		hqgologgerlevels.LevelError,
		hqgologgerlevels.LevelInfo,
		hqgologgerlevels.LevelWarn,
		hqgologgerlevels.LevelDebug,
	}

	for _, f := range factories() {
		for _, level := range colored {
			c := f.new()
			got := c.Colorize("LBL", level)

			assert.Contains(t, got, "LBL", "%s: original text preserved", f.name)
			assert.Contains(t, got, "\x1b[", "%s: ANSI escape present for level %s", f.name, level)
			assert.NotEqual(t, "LBL", got, "%s: level %s should be colorized", f.name, level)
		}
	}
}

func TestColorizeLeavesSilentAndInvalidUnchanged(t *testing.T) {
	t.Parallel()

	for _, f := range factories() {
		c := f.new()

		assert.Equal(t, "LBL", c.Colorize("LBL", hqgologgerlevels.LevelSilent), "%s: silent", f.name)
		assert.Equal(t, "LBL", c.Colorize("LBL", hqgologgerlevels.Level(99)), "%s: invalid", f.name)
	}
}

func TestColorizeDistinguishesLevels(t *testing.T) {
	t.Parallel()

	for _, f := range factories() {
		c := f.new()

		info := c.Colorize("X", hqgologgerlevels.LevelInfo)
		warn := c.Colorize("X", hqgologgerlevels.LevelWarn)
		debug := c.Colorize("X", hqgologgerlevels.LevelDebug)

		require.NotEqual(t, info, warn, "%s: info vs warn", f.name)
		require.NotEqual(t, info, debug, "%s: info vs debug", f.name)
		require.NotEqual(t, warn, debug, "%s: warn vs debug", f.name)
	}
}

func TestColorizeEmptyText(t *testing.T) {
	t.Parallel()

	for _, f := range factories() {
		c := f.new()
		got := c.Colorize("", hqgologgerlevels.LevelInfo)

		assert.True(t, strings.HasPrefix(got, "\x1b[") || got == "", "%s: %q", f.name, got)
	}
}
