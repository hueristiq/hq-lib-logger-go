package formatter

import (
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// Colorizer applies color to a label. Implementations must be safe for
// concurrent use.
//
// Methods:
//   - Colorize(text string, level hqgologgerlevels.Level) (colorized string): colors text.
type Colorizer interface {
	Colorize(text string, level hqgologgerlevels.Level) (colorized string)
}

// ColorizerFunc adapts a function to Colorizer. A nil function returns text
// unchanged.
type ColorizerFunc func(text string, level hqgologgerlevels.Level) (colorized string)

// Colorize calls f(text, level), or returns text when f is nil.
//
// Parameters:
//   - text (string): text to colorize.
//   - level (hqgologgerlevels.Level): record severity.
//
// Returns:
//   - colorized (string): the colorized text, or text when f is nil.
func (f ColorizerFunc) Colorize(text string, level hqgologgerlevels.Level) (colorized string) {
	if f == nil {
		return text
	}

	return f(text, level)
}

// NoOpColorizer returns text unchanged.
type NoOpColorizer struct{}

// Colorize returns text unchanged.
//
// Parameters:
//   - text (string): text to return.
//   - level (hqgologgerlevels.Level): record severity; ignored.
//
// Returns:
//   - colorized (string): the unchanged text.
func (NoOpColorizer) Colorize(text string, _ hqgologgerlevels.Level) (colorized string) {
	return text
}

var _ Colorizer = ColorizerFunc(nil)
var _ Colorizer = NoOpColorizer{}
