package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

// Event represents a log event with a severity level, message, timestamp, and optional
// metadata. It is used by the Logger to construct log messages before formatting and
// writing. The event is built using the options pattern, allowing flexible configuration
// of its fields via OptionFunc functions (see [NewEvent], [WithLevel], [WithMessage],
// [WithString], [WithValue], [WithLabel], and [WithError]). Construct events with
// [NewEvent]; a zero-value Event is also valid and can be configured directly through
// its Set methods.
//
// Fields:
//   - timestamp (time.Time): The time the log event was created, used for timestamped output.
//   - level (hqgologgerlevels.Level): The severity level of the log message, as defined in the levels
//     package (e.g., LevelInfo, LevelFatal). Lower values indicate higher severity.
//   - message (string): The primary content of the log message, describing the event or condition.
//   - metadata (map[string]any): Optional key-value pairs for additional context, such
//     as labels, errors, or system metrics. The reserved keys
//     [github.com/hueristiq/hq-lib-logger-go/formatter.LabelKey] and
//     [github.com/hueristiq/hq-lib-logger-go/formatter.ErrorKey] are rendered
//     specially by the console formatter.
type Event struct {
	timestamp time.Time
	level     hqgologgerlevels.Level
	message   string
	metadata  map[string]any
}

// SetTimestamp sets the timestamp of the log event, used for including timing information
// in formatted output. If not set, formatters or writers may use the current time.
//
// Parameters:
//   - t (time.Time): The timestamp to set for the log event.
func (e *Event) SetTimestamp(t time.Time) {
	e.timestamp = t
}

// SetLevel sets the severity level of the log event, determining its priority and filtering
// behavior in the logger.
//
// Parameters:
//   - l (hqgologgerlevels.Level): The severity level to set, from the levels package (e.g., LevelFatal).
func (e *Event) SetLevel(l hqgologgerlevels.Level) {
	e.level = l
}

// SetMessage sets the primary content of the log event, which forms the main body of the
// log message.
//
// Parameters:
//   - m (string): The log message to set.
func (e *Event) SetMessage(m string) {
	e.message = m
}

// SetValue adds a key-value pair to the log event's metadata with a value of any type.
// If the metadata map is nil, it is initialized before adding the pair. This allows
// flexibility for storing various data types, such as integers or errors, in metadata.
//
// Parameters:
//   - key (string): The metadata key.
//   - value (any): The metadata value, which can be any type.
func (e *Event) SetValue(key string, value any) {
	if e.metadata == nil {
		e.metadata = make(map[string]any)
	}

	e.metadata[key] = value
}

// SetString adds a key-value pair to the log event's metadata with a string value. If the
// metadata map is nil, it is initialized before adding the pair.
//
// Parameters:
//   - key (string): The metadata key.
//   - value (string): The metadata value.
func (e *Event) SetString(key, value string) {
	if e.metadata == nil {
		e.metadata = make(map[string]any)
	}

	e.metadata[key] = value
}

// SetLabel sets the reserved label metadata field for the log event, typically used by
// formatters to include a short identifier in the output (e.g., "[INFO]"). This is a
// convenience method that delegates to SetString with
// [github.com/hueristiq/hq-lib-logger-go/formatter.LabelKey].
//
// Parameters:
//   - label (string): The label to set in the metadata.
func (e *Event) SetLabel(label string) {
	e.SetString(hqgologgerformatter.LabelKey, label)
}

// SetError adds an error to the log event's metadata under the reserved error key. The
// error is stored as-is, and formatters are responsible for converting it to a string
// or other format (e.g., including stack traces). This is a convenience method that
// delegates to SetValue with
// [github.com/hueristiq/hq-lib-logger-go/formatter.ErrorKey].
//
// Parameters:
//   - err (error): The error to set in the metadata.
func (e *Event) SetError(err error) {
	e.SetValue(hqgologgerformatter.ErrorKey, err)
}

// Logger is the core component of the logging system, responsible for filtering, formatting,
// and writing log messages. It filters messages based on a configured severity threshold,
// uses a formatter to convert events to byte slices, and delegates output to a writer. The
// Logger is thread-safe, using a read-write mutex to protect configuration changes while
// allowing concurrent logging. It provides level-specific methods (e.g., Info, Fatal) for
// convenient logging and supports metadata via the options pattern. Construct a Logger
// with [NewLogger] (a zero-value Logger behaves identically), configure it with
// [Logger.SetLevel], [Logger.SetFormatter], and [Logger.SetWriter], and release its
// writer with [Logger.Close] when it is no longer needed. A Logger must not be copied
// after first use.
//
// Fields:
//   - mutex (sync.RWMutex): Ensures thread-safe access to configuration fields (level,
//     formatter, writer) during updates and concurrent logging.
//   - level (hqgologgerlevels.Level): The minimum severity level for logging (inclusive). Messages
//     with a higher level value (less severe) are ignored. Lower values indicate higher
//     severity (e.g., LevelFatal = 0, LevelDebug = 5).
//   - formatter (hqgologgerformatter.Formatter): The formatter to convert log events to byte slices
//     for output (e.g., JSON or plain text).
//   - writer (hqgologgerwriter.Writer): The writer to output formatted log data to destinations
//     like files or consoles.
type Logger struct {
	mutex     sync.RWMutex
	level     hqgologgerlevels.Level
	formatter hqgologgerformatter.Formatter
	writer    hqgologgerwriter.Writer
}

// SetLevel sets the minimum severity level for logging. Messages with a level greater
// than the specified level (less severe) are ignored. The method is thread-safe, using
// a mutex to protect the level field. The levels package uses lower values for higher
// severity (e.g., LevelFatal = 0, LevelDebug = 5). An invalid level (outside the range
// defined by the levels package) is rejected: the threshold is left unchanged and an
// error is returned, guaranteeing that LevelFatal events can never be filtered out by
// a failed update.
//
// Parameters:
//   - level (hqgologgerlevels.Level): The minimum severity level to log.
//
// Returns:
//   - err (error): An error wrapping
//     [github.com/hueristiq/hq-lib-logger-go/levels.ErrUnknownLevel] if level is
//     invalid; otherwise nil.
func (l *Logger) SetLevel(level hqgologgerlevels.Level) (err error) {
	if !level.IsValid() {
		return fmt.Errorf("%w (%d)", hqgologgerlevels.ErrUnknownLevel, level.Int())
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.level = level

	return nil
}

// Level returns the logger's current minimum severity threshold. The method is
// thread-safe, using a mutex to read the level field.
//
// Returns:
//   - level (hqgologgerlevels.Level): The current minimum severity level.
func (l *Logger) Level() (level hqgologgerlevels.Level) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	level = l.level

	return
}

// Enabled reports whether an event at the given severity level passes the logger's
// current threshold (that is, level <= configured level). Use it to skip expensive
// message construction when the output would be discarded anyway:
//
//	if logger.Enabled(hqgologgerlevels.LevelDebug) {
//		logger.Debug(expensiveDump())
//	}
//
// Note that Enabled only compares levels; it does not report whether a formatter
// and writer are configured. The method is thread-safe.
//
// Parameters:
//   - level (hqgologgerlevels.Level): The severity level to test.
//
// Returns:
//   - enabled (bool): True if an event at level would pass the threshold.
func (l *Logger) Enabled(level hqgologgerlevels.Level) (enabled bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	enabled = level <= l.level

	return
}

// SetFormatter sets the formatter used to convert log events to byte slices. The method
// is thread-safe, using a mutex to protect the formatter field. The formatter determines
// the output format, such as JSON, plain text, or structured logging formats.
//
// Parameters:
//   - f (hqgologgerformatter.Formatter): The formatter to use for log events.
func (l *Logger) SetFormatter(f hqgologgerformatter.Formatter) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.formatter = f
}

// SetWriter sets the writer used to output formatted log data to a destination (e.g.,
// console, file). The method is thread-safe, using a mutex to protect the writer field.
//
// Parameters:
//   - w (hqgologgerwriter.Writer): The writer to use for log output.
func (l *Logger) SetWriter(w hqgologgerwriter.Writer) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.writer = w
}

// Close closes the logger's writer, releasing any resources it holds (e.g., file
// handles or network connections). It returns nil when no writer is configured.
// Close does not prevent further logging — events logged afterwards are handled
// by the (possibly closed) writer — so call it once, when the logger is no longer
// needed. The method is thread-safe: it holds the logger's write lock through the
// writer's Close call, so an in-flight [Logger.Log] always finishes formatting
// and writing before the writer is closed.
//
// Returns:
//   - err (error): The error returned by the writer's Close, or nil if the close
//     succeeds or no writer is configured.
func (l *Logger) Close() (err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.writer == nil {
		return nil
	}

	return l.writer.Close()
}

// Fatal logs a message at LevelFatal, applying the provided options (e.g., metadata, labels).
// The message is formatted and written when a formatter and writer are configured (LevelFatal
// = 0, so it always passes the level filter). The program then exits with status code 1,
// indicating a critical failure — the exit happens even if the event could not be formatted
// or written, and because it is performed with os.Exit, deferred functions in the caller
// do not run. The method uses the options pattern for flexible configuration of the log event.
//
// Parameters:
//   - message (string): The log message describing the critical failure.
//   - opts (...OptionFunc): Optional configurations for the log event (e.g., metadata, error).
func (l *Logger) Fatal(message string, opts ...OptionFunc) {
	opts = append(opts, WithLevel(hqgologgerlevels.LevelFatal), WithMessage(message))

	l.Log(NewEvent(opts...))
}

// Print logs a message at LevelSilent, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelSilent). LevelSilent
// (value 1) is typically used for non-critical output, such as user-facing messages, and
// may be directed to stdout by writers. The method uses the options pattern for flexibility.
//
// Parameters:
//   - message (string): The log message for non-critical output.
//   - opts (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Print(message string, opts ...OptionFunc) {
	opts = append(opts, WithLevel(hqgologgerlevels.LevelSilent), WithMessage(message))

	l.Log(NewEvent(opts...))
}

// Error logs a message at LevelError, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelError). LevelError
// (value 2) indicates errors requiring attention but not program termination. The method
// uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message describing the error.
//   - opts (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Error(message string, opts ...OptionFunc) {
	opts = append(opts, WithLevel(hqgologgerlevels.LevelError), WithMessage(message))

	l.Log(NewEvent(opts...))
}

// Info logs a message at LevelInfo, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelInfo). LevelInfo
// (value 3) is used for informational messages about normal operation. The method uses
// the options pattern for flexibility.
//
// Parameters:
//   - message (string): The log message describing normal operation.
//   - opts (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Info(message string, opts ...OptionFunc) {
	opts = append(opts, WithLevel(hqgologgerlevels.LevelInfo), WithMessage(message))

	l.Log(NewEvent(opts...))
}

// Warn logs a message at LevelWarn, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelWarn). LevelWarn
// (value 4) indicates potential issues that do not halt execution. The method uses
// the options pattern for flexibility.
//
// Parameters:
//   - message (string): The log message describing a potential issue.
//   - opts (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Warn(message string, opts ...OptionFunc) {
	opts = append(opts, WithLevel(hqgologgerlevels.LevelWarn), WithMessage(message))

	l.Log(NewEvent(opts...))
}

// Debug logs a message at LevelDebug, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelDebug). LevelDebug
// (value 5) is used for detailed debugging information, typically enabled in development.
// The method uses the options pattern for flexibility.
//
// Parameters:
//   - message (string): The log message for debugging purposes.
//   - opts (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Debug(message string, opts ...OptionFunc) {
	opts = append(opts, WithLevel(hqgologgerlevels.LevelDebug), WithMessage(message))

	l.Log(NewEvent(opts...))
}

// Log processes a log event by filtering, formatting, and writing it. A nil event is a
// no-op. The event is ignored if its level is greater than the logger's threshold (less
// severe). The event's metadata is passed to the formatter unchanged; renderers apply
// their own conventions — the bundled console formatter substitutes a default label
// based on the level (e.g., "INF" for LevelInfo) when the event carries none. Formatting
// and writing happen only when both a formatter and a writer are configured; otherwise
// the event is silently dropped. Format and write errors are deliberately ignored —
// there is no meaningful recovery path inside a logger, and logging must never crash
// the application. For LevelFatal events the program exits with status code 1 regardless
// of whether the event was written, guaranteeing that Fatal never returns; the exit is
// performed with os.Exit, so deferred functions do not run. The method is thread-safe:
// it holds the logger's read lock for the whole operation — including the Format and
// Write calls — so a concurrent [Logger.Close] cannot close the writer in the middle
// of a write. Concurrent Log calls still proceed in parallel with each other, so the
// formatter and writer must be safe for concurrent use.
//
// Parameters:
//   - event (*Event): The log event to process, containing timestamp, level, message,
//     and metadata.
func (l *Logger) Log(event *Event) {
	if event == nil {
		return
	}

	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if event.level > l.level {
		return
	}

	if l.formatter != nil && l.writer != nil {
		data, err := l.formatter.Format(&hqgologgerformatter.Log{
			Timestamp: event.timestamp,
			Message:   event.message,
			Level:     event.level,
			Metadata:  event.metadata,
		})
		if err == nil {
			// The write error is intentionally discarded: logging must
			// never fail the caller, and there is nowhere to report it.
			l.writer.Write(data, event.level) //nolint:errcheck,gosec
		}
	}

	if event.level == hqgologgerlevels.LevelFatal {
		// Fatal must exit the process, and os.Exit skips all defers by
		// design — including the deferred RUnlock above.
		os.Exit(1) //nolint:gocritic
	}
}

// OptionFunc defines a function type for configuring log events using the options pattern.
// It allows flexible modification of an event's fields (e.g., level, message, metadata)
// during creation or logging. Custom OptionFunc implementations can be written directly
// against the exported [Event] methods.
//
// Parameters:
//   - event (*Event): The log event to configure.
type OptionFunc func(event *Event)

// NewEvent creates a new log event with the specified options. It initializes the event
// with the current time and applies the provided OptionFunc configurations to set the
// level, message, and metadata. The metadata map is allocated lazily, only when an option
// actually attaches metadata. Combined with [WithLevel] and [WithMessage], NewEvent enables
// emitting fully custom events through [Logger.Log].
//
// Parameters:
//   - opts (...OptionFunc): Configurations for the log event (e.g., level, message, metadata).
//
// Returns:
//   - event (*Event): A pointer to the configured log event.
func NewEvent(opts ...OptionFunc) (event *Event) {
	event = &Event{
		timestamp: time.Now(),
	}

	for _, f := range opts {
		f(event)
	}

	return
}

// WithLevel returns an OptionFunc that sets the severity level of a log event. It is
// primarily useful together with [NewEvent] and [Logger.Log] when emitting events
// directly; the level-specific methods (e.g., Info, Fatal) set the level themselves.
//
// Parameters:
//   - level (hqgologgerlevels.Level): The severity level to set.
//
// Returns:
//   - (OptionFunc): A function to configure the event's level.
func WithLevel(level hqgologgerlevels.Level) OptionFunc {
	return func(event *Event) {
		event.SetLevel(level)
	}
}

// WithMessage returns an OptionFunc that sets the message content of a log event. It is
// primarily useful together with [NewEvent] and [Logger.Log] when emitting events
// directly; the level-specific methods set the message themselves.
//
// Parameters:
//   - message (string): The log message to set.
//
// Returns:
//   - (OptionFunc): A function to configure the event's message.
func WithMessage(message string) OptionFunc {
	return func(event *Event) {
		event.SetMessage(message)
	}
}

// WithoutTimestamp returns an OptionFunc that clears a log event's timestamp,
// setting it to the zero time. Because the console formatter omits the timestamp
// for events whose timestamp is zero, this suppresses timestamp output for a
// single message even when the formatter is otherwise configured to include it.
// See also [WithoutLabel].
//
// Returns:
//   - (OptionFunc): A function that resets the event's timestamp to the zero value.
func WithoutTimestamp() OptionFunc {
	return func(event *Event) {
		var timestamp time.Time

		event.SetTimestamp(timestamp)
	}
}

// WithValue returns an OptionFunc that adds a key-value pair to a log event's
// metadata. Unlike [WithString], the value may be of any type; the console
// formatter renders it with the "%v" verb. It can be passed to level-specific
// logging methods (e.g., Info, Error) to attach structured context.
//
// Parameters:
//   - key (string): The metadata key.
//   - value (any): The metadata value, of any type.
//
// Returns:
//   - (OptionFunc): A function to configure the event's metadata with the value.
func WithValue(key string, value any) OptionFunc {
	return func(event *Event) {
		event.SetValue(key, value)
	}
}

// WithString returns an OptionFunc that adds a key-value pair with a string value to a
// log event's metadata. It can be passed to level-specific logging methods (e.g., Info,
// Error) to include custom metadata in the log event.
//
// Parameters:
//   - key (string): The metadata key.
//   - value (string): The metadata value.
//
// Returns:
//   - (OptionFunc): A function to configure the event's metadata with a string value.
func WithString(key, value string) OptionFunc {
	return func(event *Event) {
		event.SetString(key, value)
	}
}

// WithLabel returns an OptionFunc that sets the reserved label metadata field for a log
// event, typically used by formatters to include a short identifier in the output
// (e.g., "[INFO]"). It can be passed to level-specific logging methods to override the
// default label the console formatter derives from the level.
//
// Parameters:
//   - label (string): The label to set in the metadata.
//
// Returns:
//   - (OptionFunc): A function to configure the event's label.
func WithLabel(label string) OptionFunc {
	return func(event *Event) {
		event.SetLabel(label)
	}
}

// WithoutLabel returns an OptionFunc that suppresses the label for a log event
// by setting the reserved label metadata field to an empty string. Because the
// field is present (though empty), the console formatter does not substitute the
// level's default label and omits the bracketed label entirely. See also
// [WithLabel] and [WithoutTimestamp].
//
// Returns:
//   - (OptionFunc): A function that clears the event's label.
func WithoutLabel() OptionFunc {
	return func(event *Event) {
		event.SetLabel("")
	}
}

// WithError returns an OptionFunc that adds an error to a log event's metadata
// under the reserved error key. The error is stored as-is; formatters decide how to
// render it. The bundled console formatter prints it as a trailing block
// containing err.Error(), separated from the message by a blank line. It can be
// passed to level-specific logging methods to include error details.
//
// Parameters:
//   - err (error): The error to set in the metadata.
//
// Returns:
//   - (OptionFunc): A function to configure the event's error metadata.
func WithError(err error) OptionFunc {
	return func(event *Event) {
		event.SetError(err)
	}
}

// NewLogger creates and returns a new Logger instance with no formatter or writer
// configured and the severity threshold at its zero value (LevelFatal). Configure
// the logger with SetLevel, SetFormatter, and SetWriter before use: events are
// silently dropped until a formatter and a writer are set, though Fatal still exits.
// The logger is safe for concurrent use and must not be copied after first use.
// Call Close when the logger is no longer needed to release the writer's resources.
//
// Returns:
//   - logger (*Logger): A pointer to a new Logger instance.
func NewLogger() (logger *Logger) {
	logger = &Logger{}

	return
}
