package colorizer

import (
	"github.com/fatih/color"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// FatihColorizer is an implementation of the formatter.Colorizer interface that
// applies color formatting to log message labels using the github.com/fatih/color
// package. It maps log severity levels to distinct colors and styles (e.g., bold
// high-intensity red for LevelFatal) to enhance visual differentiation in console
// output. The colorizer is designed for terminal environments supporting ANSI
// escape codes, making log messages easier to scan and prioritize based on their
// severity. Construct with [NewFatihColorizer].
//
// Fields:
//   - fatal (*color.Color): The color configuration for LevelFatal messages,
//     using high-intensity red with bold styling.
//   - err (*color.Color): The color configuration for LevelError messages,
//     using high-intensity red with bold styling.
//   - info (*color.Color): The color configuration for LevelInfo messages,
//     using high-intensity blue with bold styling.
//   - warn (*color.Color): The color configuration for LevelWarn messages,
//     using high-intensity yellow with bold styling.
//   - debug (*color.Color): The color configuration for LevelDebug messages,
//     using high-intensity magenta with bold styling.
type FatihColorizer struct {
	fatal *color.Color
	err   *color.Color
	info  *color.Color
	warn  *color.Color
	debug *color.Color
}

// Colorize applies color and style formatting to the input text based on the
// provided log level, using the fatih/color package. Each severity level is mapped
// to a specific color and bold style for clarity: LevelFatal and LevelError are
// high-intensity red, LevelInfo is high-intensity blue, LevelWarn is high-intensity
// yellow, LevelDebug is high-intensity magenta, and LevelSilent or invalid levels
// return the text unchanged. The method satisfies the formatter.Colorizer interface
// and is intended for use with console formatters to enhance log output readability
// in terminal environments.
//
// Parameters:
//   - text (string): The input text to colorize, typically a log label (e.g., "INF").
//   - level (hqgologgerlevels.Level): The severity level of the log message, as defined in
//     the levels package (e.g., LevelFatal, LevelInfo).
//
// Returns:
//   - colorized (string): The input text with ANSI color and style formatting applied,
//     or the original text unchanged if the level is LevelSilent or invalid.
func (fc *FatihColorizer) Colorize(text string, level hqgologgerlevels.Level) (colorized string) {
	colorized = text

	switch level {
	case hqgologgerlevels.LevelFatal:
		colorized = fc.fatal.Sprint(text)
	case hqgologgerlevels.LevelError:
		colorized = fc.err.Sprint(text)
	case hqgologgerlevels.LevelInfo:
		colorized = fc.info.Sprint(text)
	case hqgologgerlevels.LevelWarn:
		colorized = fc.warn.Sprint(text)
	case hqgologgerlevels.LevelDebug:
		colorized = fc.debug.Sprint(text)
	case hqgologgerlevels.LevelSilent:
		// No color mapping; the text is returned unchanged.
	}

	return
}

var _ hqgologgerformatter.Colorizer = (*FatihColorizer)(nil)

// NewFatihColorizer creates and returns a new FatihColorizer instance, initialized
// with color configurations for each log level using the fatih/color package. Each
// level is assigned a high-intensity color with bold styling: red for LevelFatal
// and LevelError, blue for LevelInfo, yellow for LevelWarn, and magenta for
// LevelDebug. This factory function provides a convenient way to instantiate a
// FatihColorizer for use in logging systems that require colorized console output,
// such as with a Console formatter configured for colorization.
//
// Returns:
//   - colorizer (*FatihColorizer): A pointer to a new FatihColorizer instance with
//     pre-configured color settings for all log levels.
func NewFatihColorizer() (colorizer *FatihColorizer) {
	colorizer = &FatihColorizer{
		fatal: color.New(color.FgHiRed, color.Bold),
		err:   color.New(color.FgHiRed, color.Bold),
		info:  color.New(color.FgHiBlue, color.Bold),
		warn:  color.New(color.FgHiYellow, color.Bold),
		debug: color.New(color.FgHiMagenta, color.Bold),
	}

	return
}
