package formatter

import (
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// NoOpColorizer is a no-operation implementation of the Colorizer interface.
// It returns the input text unchanged, effectively disabling color formatting.
// This is useful in scenarios where color output is not desired, such as when
// logging to files, non-terminal outputs, or environments that do not support
// ANSI color codes (e.g., certain IDE consoles or CI pipelines). Construct with
// [NewNoOpColorizer].
type NoOpColorizer struct{}

// Colorize returns the input text without applying any color formatting, satisfying
// the Colorizer interface. This method is a no-op, meaning it does not modify the
// input text, making it suitable for contexts where plain text output is required.
//
// Parameters:
//   - text (string): The input text to be colorized (returned unchanged).
//   - level (hqgologgerlevels.Level): The severity level of the log message, as defined in
//     the levels package (e.g., LevelFatal, LevelInfo). Ignored by this implementation.
//
// Returns:
//   - colorized (string): The original input text, unchanged.
func (c *NoOpColorizer) Colorize(text string, level hqgologgerlevels.Level) (colorized string) {
	colorized = text

	return
}

// Colorizer defines an interface for applying color formatting to log messages
// based on their severity level. Implementations of this interface add visual
// distinctions (e.g., ANSI color codes) to log text, typically for console output,
// to make it easier to differentiate log messages by their severity. Implementations
// are expected to be safe for concurrent use, since formatting may happen from
// several goroutines at once.
//
// Methods:
//   - Colorize(text string, level levels.Level) (colorized string): Takes a text
//     string and a log level, returning the text with applied color formatting
//     specific to the provided level. The implementation determines how colors
//     are applied, such as using ANSI escape codes for terminal output or
//     returning the original text unchanged for non-colored output.
type Colorizer interface {
	Colorize(text string, level hqgologgerlevels.Level) (colorized string)
}

var _ Colorizer = (*NoOpColorizer)(nil)

// NewNoOpColorizer creates and returns a new instance of NoOpColorizer.
// This factory function provides a convenient way to instantiate a NoOpColorizer
// for use in logging systems that require a Colorizer implementation but do not
// need colorized output.
//
// Returns:
//   - colorizer (*NoOpColorizer): A pointer to a new NoOpColorizer instance.
func NewNoOpColorizer() (colorizer *NoOpColorizer) {
	colorizer = &NoOpColorizer{}

	return
}
