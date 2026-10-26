package logger

import (
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

// DefaultLogger is a pre-configured Logger instance for convenient logging without
// explicit instantiation. It is constructed at package initialization with the
// following default configuration:
//   - Level: LevelDebug (value 5), allowing all messages to be logged.
//   - Formatter: a Console formatter with plain labels (Colorize: false), producing
//     human-readable output in the format "[timestamp] [label] message [metadata]".
//   - Writer: a Console writer directing LevelSilent messages to stdout and other levels
//     (LevelFatal, LevelError, LevelWarn, LevelInfo, LevelDebug) to stderr, with newlines
//     appended.
//
// Package-level functions (Fatal, Print, Error, Info, Warn, Debug) delegate to
// DefaultLogger, enabling immediate logging with minimal setup. The Logger filters
// messages based on its level threshold (lower values indicate higher severity, e.g.,
// LevelFatal = 0), and exits the program with status code 1 for LevelFatal messages.
// The console formatter supplies default labels when an event carries none (e.g.,
// "INF" for LevelInfo). Users can modify DefaultLogger's configuration (e.g., level,
// formatter, writer) to customize behavior or create a new Logger instance for more
// control. The Logger is thread-safe for configuration changes and relies on the
// formatter and writer for their own thread-safety. Reassigning DefaultLogger itself
// is not synchronized; do it during startup, before any goroutine may call the
// package-level functions.
var DefaultLogger = newDefaultLogger()

// newDefaultLogger builds the DefaultLogger with its default level, formatter, and
// writer. Keeping construction in a function avoids an init() while guaranteeing the
// variable is fully initialized before any package-level function can run.
//
// Returns:
//   - logger (*Logger): the configured default Logger.
func newDefaultLogger() (logger *Logger) {
	logger = NewLogger()

	// LevelDebug is a valid level, so SetLevel cannot fail here.
	_ = logger.SetLevel(hqgologgerlevels.LevelDebug)

	logger.SetFormatter(hqgologgerformatter.NewConsoleFormatter(hqgologgerformatter.DefaultConsoleFormatterConfig()))
	logger.SetWriter(hqgologgerwriter.NewConsoleWriter(hqgologgerwriter.DefaultConsoleWriterConfig()))

	return
}

// Fatal logs through [DefaultLogger]; see [Logger.Fatal] for the exit and
// option semantics.
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func Fatal(message string, opts ...OptionFunc) {
	DefaultLogger.Fatal(message, opts...)
}

// Print logs through [DefaultLogger]; see [Logger.Print].
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func Print(message string, opts ...OptionFunc) {
	DefaultLogger.Print(message, opts...)
}

// Error logs through [DefaultLogger]; see [Logger.Error].
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func Error(message string, opts ...OptionFunc) {
	DefaultLogger.Error(message, opts...)
}

// Info logs through [DefaultLogger]; see [Logger.Info].
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func Info(message string, opts ...OptionFunc) {
	DefaultLogger.Info(message, opts...)
}

// Warn logs through [DefaultLogger]; see [Logger.Warn].
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func Warn(message string, opts ...OptionFunc) {
	DefaultLogger.Warn(message, opts...)
}

// Debug logs through [DefaultLogger]; see [Logger.Debug].
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func Debug(message string, opts ...OptionFunc) {
	DefaultLogger.Debug(message, opts...)
}
