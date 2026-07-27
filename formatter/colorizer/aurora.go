package colorizer

import (
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/logrusorgru/aurora/v4"
)

// AuroraColorizer is an implementation of the formatter.Colorizer interface that
// applies color formatting to log message labels using the aurora package. It maps
// log severity levels to distinct colors and styles (e.g., bold red for LevelFatal)
// to enhance visual differentiation in console output. The colorizer is designed
// for terminal environments supporting ANSI escape codes, making log messages easier
// to scan and prioritize based on their severity.
//
// Fields:
//   - au (*aurora.Aurora): The aurora instance used for applying color and style
//     formatting to text. Configured to enable colors by default.
type AuroraColorizer struct {
	au *aurora.Aurora
}

// Colorize applies color and style formatting to the input text based on the provided
// log level, using the aurora package. Each severity level is mapped to a specific
// color and bold style for clarity: LevelFatal and LevelError are bright red,
// LevelInfo is bright blue, LevelWarn is bright yellow, LevelDebug is bright magenta,
// and LevelSilent or invalid levels return the text unchanged. The method satisfies
// the formatter.Colorizer interface and is intended for use with console formatters
// to enhance log output readability in terminal environments.
//
// Parameters:
//   - text (string): The input text to colorize, typically a log label (e.g., "INF").
//   - level (hqgologgerlevels.Level): The severity level of the log message, as defined in
//     the levels package (e.g., LevelFatal, LevelInfo).
//
// Returns:
//   - colorized (string): The input text with ANSI color and style formatting applied,
//     or the original text unchanged if the level is LevelSilent or invalid.
func (c *AuroraColorizer) Colorize(text string, level hqgologgerlevels.Level) (colorized string) {
	colorized = text

	switch level {
	case hqgologgerlevels.LevelFatal:
		colorized = c.au.BrightRed(text).Bold().String()
	case hqgologgerlevels.LevelError:
		colorized = c.au.BrightRed(text).Bold().String()
	case hqgologgerlevels.LevelInfo:
		colorized = c.au.BrightBlue(text).Bold().String()
	case hqgologgerlevels.LevelWarn:
		colorized = c.au.BrightYellow(text).Bold().String()
	case hqgologgerlevels.LevelDebug:
		colorized = c.au.BrightMagenta(text).Bold().String()
	case hqgologgerlevels.LevelSilent:
		// No color mapping; the text is returned unchanged.
	}

	return
}

var _ hqgologgerformatter.Colorizer = (*AuroraColorizer)(nil)

// NewAuroraColorizer creates and returns a new AuroraColorizer instance, initialized
// with an aurora instance configured to enable ANSI color output. This factory function
// provides a convenient way to instantiate an AuroraColorizer for use in logging
// systems that require colorized console output, such as with a Console formatter
// configured for colorization.
//
// Returns:
//   - colorizer (*AuroraColorizer): A pointer to a new AuroraColorizer instance with
//     colors enabled.
func NewAuroraColorizer() (colorizer *AuroraColorizer) {
	colorizer = &AuroraColorizer{
		au: aurora.New(aurora.WithColors(true)),
	}

	return
}
