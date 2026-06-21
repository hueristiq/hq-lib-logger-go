package logger

import (
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

// DefaultLogger is a pre-configured Logger instance for convenient logging without
// explicit instantiation. It is initialized in the init() function with the following
// default configuration:
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
// LevelFatal = 0), adds default labels if none are provided (e.g., "INF" for LevelInfo),
// and exits the program with status code 1 for LevelFatal messages. Users can modify
// DefaultLogger’s configuration (e.g., level, formatter, writer) to customize behavior
// or create a new Logger instance for more control. The Logger is thread-safe for
// configuration changes and relies on the formatter and writer for their own thread-safety.
var DefaultLogger *Logger

func init() {
	DefaultLogger = NewLogger()

	DefaultLogger.SetLevel(hqgologgerlevels.LevelDebug)
	DefaultLogger.SetFormatter(hqgologgerformatter.NewConsoleFormatter(hqgologgerformatter.DefaultConsoleConfig()))
	DefaultLogger.SetWriter(hqgologgerwriter.NewConsoleWriter(hqgologgerwriter.DefaultConsoleWriterConfig()))
}

// Fatal logs a message at LevelFatal using DefaultLogger, applying the provided options
// (e.g., metadata, labels). The message is formatted and written if the logger’s threshold
// allows (LevelFatal = 0, so it is always logged unless the formatter or writer is nil).
// After writing, the program exits with status code 1, indicating a critical failure.
// The method uses the options pattern for flexible configuration of the log event.
//
// Parameters:
//   - message (string): The log message describing the critical failure.
//   - ofs (...OptionFunc): Optional configurations for the log event (e.g., metadata, error).
func Fatal(message string, ofs ...OptionFunc) {
	DefaultLogger.Fatal(message, ofs...)
}

// Print logs a message at LevelSilent using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger’s threshold allows (level <= LevelSilent).
// LevelSilent (value 1) is typically used for non-critical output, such as user-facing messages,
// and is directed to stdout by the default Console writer. The method uses the options pattern
// for flexible configuration.
//
// Parameters:
//   - message (string): The log message for non-critical output.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func Print(message string, ofs ...OptionFunc) {
	DefaultLogger.Print(message, ofs...)
}

// Error logs a message at LevelError using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger’s threshold allows (level <= LevelError).
// LevelError (value 2) indicates errors requiring attention but not program termination.
// The method uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message describing the error.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func Error(message string, ofs ...OptionFunc) {
	DefaultLogger.Error(message, ofs...)
}

// Info logs a message at LevelInfo using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger’s threshold allows (level <= LevelInfo).
// LevelInfo (value 3) is used for informational messages about normal operation. The method
// uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message describing normal operation.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func Info(message string, ofs ...OptionFunc) {
	DefaultLogger.Info(message, ofs...)
}

// Warn logs a message at LevelWarn using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger’s threshold allows (level <= LevelWarn).
// LevelWarn (value 4) indicates potential issues that do not halt execution. The method
// uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message describing a potential issue.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func Warn(message string, ofs ...OptionFunc) {
	DefaultLogger.Warn(message, ofs...)
}

// Debug logs a message at LevelDebug using DefaultLogger, applying the provided options.
// The message is formatted and written if the logger’s threshold allows (level <= LevelDebug).
// LevelDebug (value 5) is used for detailed debugging information, typically enabled in
// development environments. The method uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message for debugging purposes.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func Debug(message string, ofs ...OptionFunc) {
	DefaultLogger.Debug(message, ofs...)
}
