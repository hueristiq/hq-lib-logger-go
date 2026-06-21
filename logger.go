package logger

import (
	"os"
	"strings"
	"sync"
	"time"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

// _Event represents a log event with a severity level, message, timestamp, and optional metadata.
// It is used internally by the Logger to construct log messages before formatting and writing.
// The event is built using the options pattern, allowing flexible configuration of its fields
// via OptionFunc functions. The timestamp is set by formatters or writers if not explicitly
// configured.
//
// Fields:
//   - timestamp (time.Time): The time the log event was created, used for timestamped output.
//   - level (hqgologgerlevels.Level): The severity level of the log message, as defined in the levels
//     package (e.g., LevelInfo, LevelFatal). Lower values indicate higher severity.
//   - message (string): The primary content of the log message, describing the event or condition.
//   - metadata (map[string]any): Optional key-value pairs for additional context, such
//     as labels, errors, or system metrics. The "label" key is used for formatted output, and
//     the "error" key is used for error details.
type _Event struct {
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
func (e *_Event) SetTimestamp(t time.Time) {
	e.timestamp = t
}

// SetLevel sets the severity level of the log event, determining its priority and filtering
// behavior in the logger.
//
// Parameters:
//   - l (hqgologgerlevels.Level): The severity level to set, from the levels package (e.g., LevelFatal).
func (e *_Event) SetLevel(l hqgologgerlevels.Level) {
	e.level = l
}

// SetMessage sets the primary content of the log event, which forms the main body of the
// log message.
//
// Parameters:
//   - m (string): The log message to set.
func (e *_Event) SetMessage(m string) {
	e.message = m
}

// SetValue adds a key-value pair to the log event's metadata with a value of any type.
// If the metadata map is nil, it is initialized before adding the pair. This allows
// flexibility for storing various data types, such as integers or errors, in metadata.
//
// Parameters:
//   - key (string): The metadata key.
//   - value (any): The metadata value, which can be any type.
func (e *_Event) SetValue(key string, value any) {
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
func (e *_Event) SetString(key, value string) {
	if e.metadata == nil {
		e.metadata = make(map[string]any)
	}

	e.metadata[key] = value
}

// SetLabel sets the "label" metadata field for the log event, typically used by formatters
// to include a short identifier in the output (e.g., "[INFO]"). This is a convenience method
// that delegates to SetString with the key "label".
//
// Parameters:
//   - label (string): The label to set in the metadata.
func (e *_Event) SetLabel(label string) {
	e.SetString("label", label)
}

// SetError adds an error to the log event's metadata under the "error" key. The error is
// stored as-is, and formatters are responsible for converting it to a string or other format
// (e.g., including stack traces). This is a convenience method that delegates to SetValue.
//
// Parameters:
//   - err (error): The error to set in the metadata.
func (e *_Event) SetError(err error) {
	e.SetValue("error", err)
}

// Logger is the core component of the logging system, responsible for filtering, formatting,
// and writing log messages. It filters messages based on a configured severity threshold,
// uses a formatter to convert events to byte slices, and delegates output to a writer. The
// Logger is thread-safe, using a read-write mutex to protect configuration changes while
// allowing concurrent logging. It provides level-specific methods (e.g., Info, Fatal) for
// convenient logging and supports metadata via the options pattern.
//
// Fields:
//   - mutex (*sync.RWMutex): Ensures thread-safe access to configuration fields (level,
//     formatter, writer) during updates and concurrent logging.
//   - level (hqgologgerlevels.Level): The minimum severity level for logging (inclusive). Messages
//     with a higher level value (less severe) are ignored. Lower values indicate higher
//     severity (e.g., LevelFatal = 0, LevelDebug = 5).
//   - formatter (hqgologgerformatter.Formatter): The formatter to convert log events to byte slices
//     for output (e.g., JSON or plain text).
//   - writer (hqgologgerwriter.Writer): The writer to output formatted log data to destinations
//     like files or consoles.
type Logger struct {
	mutex     *sync.RWMutex
	level     hqgologgerlevels.Level
	formatter hqgologgerformatter.Formatter
	writer    hqgologgerwriter.Writer
}

// SetLevel sets the minimum severity level for logging. Messages with a level greater
// than the specified level (less severe) are ignored. The method is thread-safe, using
// a mutex to protect the level field. The levels package uses lower values for higher
// severity (e.g., LevelFatal = 0, LevelDebug = 5).
//
// Parameters:
//   - level (hqgologgerlevels.Level): The minimum severity level to log.
func (l *Logger) SetLevel(level hqgologgerlevels.Level) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.level = level
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

// Fatal logs a message at LevelFatal, applying the provided options (e.g., metadata, labels).
// The message is formatted and written if the logger's threshold allows (LevelFatal = 0,
// so it is always logged unless formatter or writer is nil). After writing, the program
// exits with status code 1, indicating a critical failure. The method uses the options
// pattern for flexible configuration of the log event.
//
// Parameters:
//   - message (string): The log message describing the critical failure.
//   - ofs (...OptionFunc): Optional configurations for the log event (e.g., metadata, error).
func (l *Logger) Fatal(message string, ofs ...OptionFunc) {
	ofs = append(ofs, _WithLevel(hqgologgerlevels.LevelFatal), _WithMessage(message))

	l.Log(_NewEvent(ofs...))
}

// Print logs a message at LevelSilent, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelSilent). LevelSilent
// (value 1) is typically used for non-critical output, such as user-facing messages, and
// may be directed to stdout by writers. The method uses the options pattern for flexibility.
//
// Parameters:
//   - message (string): The log message for non-critical output.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Print(message string, ofs ...OptionFunc) {
	ofs = append(ofs, _WithLevel(hqgologgerlevels.LevelSilent), _WithMessage(message))

	l.Log(_NewEvent(ofs...))
}

// Error logs a message at LevelError, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelError). LevelError
// (value 2) indicates errors requiring attention but not program termination. The method
// uses the options pattern for flexible configuration.
//
// Parameters:
//   - message (string): The log message describing the error.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Error(message string, ofs ...OptionFunc) {
	ofs = append(ofs, _WithLevel(hqgologgerlevels.LevelError), _WithMessage(message))

	l.Log(_NewEvent(ofs...))
}

// Info logs a message at LevelInfo, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelInfo). LevelInfo
// (value 3) is used for informational messages about normal operation. The method uses
// the options pattern for flexibility.
//
// Parameters:
//   - message (string): The log message describing normal operation.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Info(message string, ofs ...OptionFunc) {
	ofs = append(ofs, _WithLevel(hqgologgerlevels.LevelInfo), _WithMessage(message))

	l.Log(_NewEvent(ofs...))
}

// Warn logs a message at LevelWarn, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelWarn). LevelWarn
// (value 4) indicates potential issues that do not halt execution. The method uses
// the options pattern for flexibility.
//
// Parameters:
//   - message (string): The log message describing a potential issue.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Warn(message string, ofs ...OptionFunc) {
	ofs = append(ofs, _WithLevel(hqgologgerlevels.LevelWarn), _WithMessage(message))

	l.Log(_NewEvent(ofs...))
}

// Debug logs a message at LevelDebug, applying the provided options. The message is
// formatted and written if the logger's threshold allows (level <= LevelDebug). LevelDebug
// (value 5) is used for detailed debugging information, typically enabled in development.
// The method uses the options pattern for flexibility.
//
// Parameters:
//   - message (string): The log message for debugging purposes.
//   - ofs (...OptionFunc): Optional configurations for the log event.
func (l *Logger) Debug(message string, ofs ...OptionFunc) {
	ofs = append(ofs, _WithLevel(hqgologgerlevels.LevelDebug), _WithMessage(message))

	l.Log(_NewEvent(ofs...))
}

// Log processes a log event by filtering, formatting, and writing it. The event is ignored
// if its level is greater than the logger's threshold (less severe). If no "label" is
// provided in the event's metadata, a default label is added based on the level (e.g., "INF"
// for LevelInfo). The message is trimmed of trailing newlines before formatting. If the
// formatter or writer is nil, or if formatting fails, the event is silently ignored. For
// LevelFatal events, the program exits with status code 1 after writing. The method is
// thread-safe for reading configuration but relies on the formatter and writer for their
// own thread-safety.
//
// Parameters:
//   - event (*_Event): The log event to process, containing timestamp, level, message,
//     and metadata.
func (l *Logger) Log(event *_Event) {
	l.mutex.RLock()
	formatter, writer, level := l.formatter, l.writer, l.level
	l.mutex.RUnlock()

	if formatter == nil || writer == nil || event.level > level {
		return
	}

	if _, ok := event.metadata["label"]; !ok {
		if label, ok := _defaultLabels[event.level]; ok {
			event.metadata["label"] = label
		}
	}

	event.message = strings.TrimSuffix(event.message, "\n")

	data, err := formatter.Format(&hqgologgerformatter.Log{
		Timestamp: event.timestamp,
		Message:   event.message,
		Level:     event.level,
		Metadata:  event.metadata,
	})
	if err != nil {
		return
	}

	writer.Write(data, event.level)

	if event.level == hqgologgerlevels.LevelFatal {
		os.Exit(1)
	}
}

// _defaultLabels maps each severity level to the short label applied to a log
// event when none is supplied via WithLabel. LevelSilent has no default label,
// so events logged via Print render without a bracketed tag unless one is set.
var _defaultLabels = map[hqgologgerlevels.Level]string{
	hqgologgerlevels.LevelFatal: "FTL",
	hqgologgerlevels.LevelError: "ERR",
	hqgologgerlevels.LevelInfo:  "INF",
	hqgologgerlevels.LevelWarn:  "WRN",
	hqgologgerlevels.LevelDebug: "DBG",
}

// OptionFunc defines a function type for configuring log events using the options pattern.
// It allows flexible modification of an event’s fields (e.g., level, message, metadata)
// during creation or logging.
//
// Parameters:
//   - event (*_Event): The log event to configure.
type OptionFunc func(event *_Event)

// _NewEvent creates a new log event with the specified options. It initializes the event
// with an empty metadata map and applies the provided OptionFunc configurations to set
// the timestamp, level, message, and metadata. The timestamp is typically set by formatters
// or writers if not explicitly configured.
//
// Parameters:
//   - ofs (...OptionFunc): Configurations for the log event (e.g., level, message, metadata).
//
// Returns:
//   - event (*_Event): A pointer to the configured log event.
func _NewEvent(ofs ...OptionFunc) (event *_Event) {
	event = &_Event{
		timestamp: time.Now(),
		metadata:  make(map[string]any),
	}

	for _, f := range ofs {
		f(event)
	}

	return
}

// _WithLevel returns an OptionFunc that sets the severity level of a log event. This is
// an internal helper used by level-specific logging methods (e.g., Info, Fatal) to configure
// the event’s level.
//
// Parameters:
//   - level (hqgologgerlevels.Level): The severity level to set.
//
// Returns:
//   - (OptionFunc): A function to configure the event’s level.
func _WithLevel(level hqgologgerlevels.Level) OptionFunc {
	return func(event *_Event) {
		event.SetLevel(level)
	}
}

// _WithMessage returns an OptionFunc that sets the message content of a log event. This
// is an internal helper used by level-specific logging methods to configure the event’s
// message.
//
// Parameters:
//   - message (string): The log message to set.
//
// Returns:
//   - (OptionFunc): A function to configure the event’s message.
func _WithMessage(message string) OptionFunc {
	return func(event *_Event) {
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
	return func(event *_Event) {
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
	return func(event *_Event) {
		event.SetValue(key, value)
	}
}

// WithString returns an OptionFunc that adds a key-value pair with a string value to a
// log event’s metadata. It can be passed to level-specific logging methods (e.g., Info,
// Error) to include custom metadata in the log event.
//
// Parameters:
//   - key (string): The metadata key.
//   - value (string): The metadata value.
//
// Returns:
//   - (OptionFunc): A function to configure the event’s metadata with a string value.
func WithString(key, value string) OptionFunc {
	return func(event *_Event) {
		event.SetString(key, value)
	}
}

// WithLabel returns an OptionFunc that sets the "label" metadata field for a log event,
// typically used by formatters to include a short identifier in the output (e.g., "[INFO]").
// It can be passed to level-specific logging methods to override the default label.
//
// Parameters:
//   - label (string): The label to set in the metadata.
//
// Returns:
//   - (OptionFunc): A function to configure the event’s label.
func WithLabel(label string) OptionFunc {
	return func(event *_Event) {
		event.SetLabel(label)
	}
}

// WithoutLabel returns an OptionFunc that suppresses the label for a log event
// by setting the "label" metadata field to an empty string. Because the field is
// present (though empty), [Logger.Log] does not substitute the level's default
// label, and the console formatter omits the bracketed label entirely. See also
// [WithLabel] and [WithoutTimestamp].
//
// Returns:
//   - (OptionFunc): A function that clears the event's label.
func WithoutLabel() OptionFunc {
	return func(event *_Event) {
		event.SetLabel("")
	}
}

// WithError returns an OptionFunc that adds an error to a log event’s metadata
// under the "error" key. The error is stored as-is; formatters decide how to
// render it. The bundled console formatter prints it as a trailing block
// containing err.Error(), separated from the message by a blank line. It can be
// passed to level-specific logging methods to include error details.
//
// Parameters:
//   - err (error): The error to set in the metadata.
//
// Returns:
//   - (OptionFunc): A function to configure the event’s error metadata.
func WithError(err error) OptionFunc {
	return func(event *_Event) {
		event.SetError(err)
	}
}

// NewLogger creates and returns a new Logger instance with a read-write mutex for
// thread-safe configuration but no formatter, writer, or level set. Users must configure
// the logger with a level, formatter, and writer before use to avoid silent failures
// during logging. The logger is ready for customization and use in a logging system.
//
// Returns:
//   - logger (*Logger): A pointer to a new Logger instance with a mutex initialized.
func NewLogger() (logger *Logger) {
	logger = &Logger{
		mutex: &sync.RWMutex{},
	}

	return
}
