package colorizer

import (
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// ANSIColorizer is an implementation of the formatter.Colorizer interface
// that applies color formatting to log message labels using raw ANSI SGR
// escape sequences. It maps log severity levels to high-intensity colors
// with bold styling (matching the SGR sequences the former
// fatih/color-backed colorizer emitted): LevelFatal and LevelError are
// high-intensity red, LevelInfo is high-intensity blue, LevelWarn is
// high-intensity yellow, and LevelDebug is high-intensity magenta;
// LevelSilent and invalid levels return the text unchanged. The colorizer
// is stateless, so it is safe for concurrent use. Construct with
// [NewANSIColorizer].
type ANSIColorizer struct{}

// Colorize wraps the input text in the SGR escape sequence of the given
// log level, or returns it unchanged when the level is LevelSilent or
// invalid.
//
// Parameters:
//   - text (string): The input text to colorize, typically a log label (e.g., "INF").
//   - level (hqgologgerlevels.Level): The severity level of the log message.
//
// Returns:
//   - colorized (string): The input text with ANSI color and style
//     formatting applied, or the original text unchanged.
func (c *ANSIColorizer) Colorize(text string, level hqgologgerlevels.Level) (colorized string) {
	colorized = text

	switch level {
	case hqgologgerlevels.LevelFatal:
		colorized = sgr(text, "1;91")
	case hqgologgerlevels.LevelError:
		colorized = sgr(text, "1;91")
	case hqgologgerlevels.LevelInfo:
		colorized = sgr(text, "1;94")
	case hqgologgerlevels.LevelWarn:
		colorized = sgr(text, "1;93")
	case hqgologgerlevels.LevelDebug:
		colorized = sgr(text, "1;95")
	case hqgologgerlevels.LevelSilent:
		// No color mapping; the text is returned unchanged.
	}

	return
}

// sgr wraps text in the given SGR parameter sequence.
func sgr(text, params string) string {
	return "\x1b[" + params + "m" + text + "\x1b[0m"
}

// Compile-time guard ensuring [ANSIColorizer] satisfies
// [hqgologgerformatter.Colorizer].
var _ hqgologgerformatter.Colorizer = (*ANSIColorizer)(nil)

// NewANSIColorizer creates and returns a new ANSIColorizer instance.
//
// Returns:
//   - colorizer (*ANSIColorizer): A pointer to a new ANSIColorizer.
func NewANSIColorizer() (colorizer *ANSIColorizer) {
	colorizer = &ANSIColorizer{}

	return
}
