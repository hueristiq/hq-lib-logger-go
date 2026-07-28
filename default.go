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
//   - Formatter: A Console formatter with colorized labels (Colorize: true), producing
//     human-readable output in the format "[timestamp] [label] message [metadata]".
//   - Writer: A Console writer directing LevelSilent messages to stdout and other levels
//     (LevelFatal, LevelError, LevelInfo, LevelWarn, LevelDebug) to stderr, with newlines
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
//   - logger (*Logger): The configured default Logger.
func newDefaultLogger() (logger *Logger) {
	logger = NewLogger()

	// LevelDebug is a valid level, so SetLevel cannot fail here.
	_ = logger.SetLevel(hqgologgerlevels.LevelDebug)

	logger.SetFormatter(hqgologgerformatter.NewConsoleFormatter(hqgologgerformatter.DefaultConsoleFormatterConfig()))
	logger.SetWriter(hqgologgerwriter.NewConsoleWriter(hqgologgerwriter.DefaultConsoleWriterConfig()))

	return
}

// Fatal logs a message at LevelFatal using DefaultLogger, applying the provided options
// (e.g., metadata, labels). The message is formatted and written when a formatter and
// writer are configured (LevelFatal = 0, so it always passes the level filter). The
// program then exits with status code 1, indicating a critical failure — the exit happens
// even if the event could not be formatted or written, and because it is performed with
// os.Exit, deferred functions in the caller do not run. The function uses the options
// pattern for flexible configuration of the log event.
//
// Parameters:
//   - message (string): The log message describing the critical failure.
//   - opts (...OptionFunc): Optional configurations for the log event (e.g., metadata, error).
func Fatal(message string, opts ...OptionFunc) {
	DefaultLogger.Fatal(message, opts...)
}

// Print logs a message at LevelSilent using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger's threshold allows (level <= LevelSilent).
// LevelSilent (value 1) is typically used for non-critical output, such as user-facing messages,
// and is directed to stdout by the default Console writer. The function uses the options pattern
// for flexible configuration.
//
// Parameters:
//   - message (string): The log message for non-critical output.
//   - opts (...OptionFunc): Optional configurations for the log event.
func Print(message string, opts ...OptionFunc) {
	DefaultLogger.Print(message, opts...)
}

// Error logs a message at LevelError using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger's threshold allows (level <= LevelError).
// LevelError (value 2) indicates errors requiring attention but not program termination.
// The function uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message describing the error.
//   - opts (...OptionFunc): Optional configurations for the log event.
func Error(message string, opts ...OptionFunc) {
	DefaultLogger.Error(message, opts...)
}

// Info logs a message at LevelInfo using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger's threshold allows (level <= LevelInfo).
// LevelInfo (value 3) is used for informational messages about normal operation. The function
// uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message describing normal operation.
//   - opts (...OptionFunc): Optional configurations for the log event.
func Info(message string, opts ...OptionFunc) {
	DefaultLogger.Info(message, opts...)
}

// Warn logs a message at LevelWarn using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger's threshold allows (level <= LevelWarn).
// LevelWarn (value 4) indicates potential issues that do not halt execution. The function
// uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message describing a potential issue.
//   - opts (...OptionFunc): Optional configurations for the log event.
func Warn(message string, opts ...OptionFunc) {
	DefaultLogger.Warn(message, opts...)
}

// Debug logs a message at LevelDebug using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger's threshold allows (level <= LevelDebug).
// LevelDebug (value 5) is used for detailed debugging information, typically enabled in
// development environments. The function uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message for debugging purposes.
//   - opts (...OptionFunc): Optional configurations for the log event.
func Debug(message string, opts ...OptionFunc) {
	DefaultLogger.Debug(message, opts...)
}
