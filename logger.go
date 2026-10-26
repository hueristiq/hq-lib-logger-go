package logger

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
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
// [NewEvent]; a zero-value Event can be configured directly through its Set methods,
// but its level must be set with [Event.SetLevel] (or [NewEvent]) before [Logger.Log]
// will emit it.
//
// Fields:
//   - timestamp (time.Time): time the log event was created.
//   - level (hqgologgerlevels.Level): severity level of the log message.
//   - levelSet (bool): whether the level was set explicitly; [Logger.Log] drops
//     events without it.
//   - message (string): primary content of the log message.
//   - metadata (map[string]any): optional key-value context; lazily allocated
//     by the Set* helpers.
type Event struct {
	timestamp time.Time
	level     hqgologgerlevels.Level
	levelSet  bool
	message   string
	metadata  map[string]any
}

// SetTimestamp sets the timestamp of the log event, used for including timing information
// in formatted output. An unset timestamp is omitted by the bundled console formatter;
// no writer substitutes the current time.
//
// Parameters:
//   - t (time.Time): the timestamp to set for the log event.
func (e *Event) SetTimestamp(t time.Time) {
	e.timestamp = t
}

// SetLevel sets the severity level of the log event, determining its priority and filtering
// behavior in the logger. Setting a level also marks the event as having an explicit level;
// [Logger.Log] drops events whose level was never set.
//
// Parameters:
//   - l (hqgologgerlevels.Level): the severity level to set, from the levels package (e.g., LevelFatal).
func (e *Event) SetLevel(l hqgologgerlevels.Level) {
	e.level = l
	e.levelSet = true
}

// SetMessage sets the primary content of the log event, which forms the main body of the
// log message.
//
// Parameters:
//   - m (string): the log message to set.
func (e *Event) SetMessage(m string) {
	e.message = m
}

// SetValue adds a key-value pair to the log event's metadata with a value of any type.
// If the metadata map is nil, it is initialized before adding the pair. This allows
// flexibility for storing various data types, such as integers or errors, in metadata.
//
// Parameters:
//   - key (string): the metadata key.
//   - value (any): the metadata value, which can be any type.
func (e *Event) SetValue(key string, value any) {
	if e.metadata == nil {
		e.metadata = make(map[string]any)
	}

	e.metadata[key] = value
}

// SetString adds a key-value pair to the log event's metadata with a string value.
// It delegates to [Event.SetValue].
//
// Parameters:
//   - key (string): the metadata key.
//   - value (string): the metadata value.
func (e *Event) SetString(key, value string) {
	e.SetValue(key, value)
}

// SetLabel sets the reserved label metadata field for the log event, typically used by
// formatters to include a short identifier in the output (e.g., "[INFO]"). This is a
// convenience method that delegates to SetString with
// [github.com/hueristiq/hq-lib-logger-go/formatter.LabelKey].
//
// Parameters:
//   - label (string): the label to set in the metadata.
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
//   - err (error): the error to set in the metadata.
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
//   - mutex (sync.RWMutex): guards configuration and Close state.
//   - level (hqgologgerlevels.Level): severity threshold.
//   - threshold (atomic.Int64): lock-free mirror of level for Log and Enabled.
//   - formatter (hqgologgerformatter.Formatter): converts events to bytes.
//   - writer (hqgologgerwriter.Writer): delivers formatted output.
//   - closed (bool): whether Close has run.
//   - closeErr (error): result of the first Close.
type Logger struct {
	mutex     sync.RWMutex
	level     hqgologgerlevels.Level
	threshold atomic.Int64
	formatter hqgologgerformatter.Formatter
	writer    hqgologgerwriter.Writer
	closed    bool
	closeErr  error
}

// SetLevel sets the minimum severity level for logging. Messages with a level greater
// than the specified level (less severe) are ignored. The method is thread-safe, using
// a mutex to protect the level field, and also publishes the threshold to the lock-free
// atomic mirror read by [Logger.Log] and [Logger.Enabled]. The levels package uses
// lower values for higher severity (e.g., LevelFatal = 0, LevelDebug = 5). An invalid
// level (outside the range defined by the levels package) is rejected: the threshold
// is left unchanged and an error is returned, guaranteeing that LevelFatal events can
// never be filtered out by a failed update.
//
// Parameters:
//   - level (hqgologgerlevels.Level): the minimum severity level to log.
//
// Returns:
//   - err (error): an error wrapping
//     [github.com/hueristiq/hq-lib-logger-go/levels.ErrUnknownLevel] if level is
//     invalid; otherwise nil.
func (l *Logger) SetLevel(level hqgologgerlevels.Level) (err error) {
	if !level.IsValid() {
		return fmt.Errorf("%w (%d)", hqgologgerlevels.ErrUnknownLevel, level.Int())
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.level = level

	l.threshold.Store(int64(level))

	return nil
}

// Level returns the logger's current minimum severity threshold. The method is
// thread-safe, using a mutex to read the level field.
//
// Returns:
//   - level (hqgologgerlevels.Level): the current minimum severity level.
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
// and writer are configured. The method is thread-safe and lock-free: the threshold
// is read from the atomic mirror maintained by [Logger.SetLevel].
//
// Parameters:
//   - level (hqgologgerlevels.Level): the severity level to test.
//
// Returns:
//   - enabled (bool): true if an event at level would pass the threshold.
func (l *Logger) Enabled(level hqgologgerlevels.Level) (enabled bool) {
	enabled = level <= hqgologgerlevels.Level(l.threshold.Load())

	return
}

// SetFormatter sets the formatter used to convert log events to byte slices. The method
// is thread-safe, using a mutex to protect the formatter field. The formatter determines
// the output format, such as JSON, plain text, or structured logging formats.
//
// Parameters:
//   - f (hqgologgerformatter.Formatter): the formatter to use for log events.
func (l *Logger) SetFormatter(f hqgologgerformatter.Formatter) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.formatter = f
}

// SetWriter sets the writer used to output formatted log data to a destination (e.g.,
// console, file). The method is thread-safe, using a mutex to protect the writer
// field. SetWriter does not close the previously configured writer; the caller owns
// it. Setting a writer also re-arms [Logger.Close], so a writer installed after a
// Close can be closed by a later Close.
//
// Parameters:
//   - w (hqgologgerwriter.Writer): the writer to use for log output.
func (l *Logger) SetWriter(w hqgologgerwriter.Writer) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.writer = w
	l.closed = false
	l.closeErr = nil
}

// Close closes the logger's writer, releasing any resources it holds (e.g., file
// handles or network connections). It returns nil when no writer is configured.
// Repeated calls are safe and return the first call's error. Close does not prevent
// further logging — events logged afterwards are handled by the (possibly closed)
// writer — so call it when the logger is no longer needed. The method is
// thread-safe: it holds the logger's write lock through the writer's Close call, so
// an in-flight [Logger.Log] always finishes formatting and writing before the writer
// is closed.
//
// Returns:
//   - err (error): the error returned by the writer's Close, or nil if the close
//     succeeds or no writer is configured.
func (l *Logger) Close() (err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.closed {
		return l.closeErr
	}

	l.closed = true

	if l.writer == nil {
		return nil
	}

	l.closeErr = l.writer.Close()

	return l.closeErr
}

// Fatal logs at [LevelFatal] and then exits with status code 1, regardless of
// whether the event was formatted or written. Options cannot change the level.
// See [Logger.Log] for filtering and formatting.
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func (l *Logger) Fatal(message string, opts ...OptionFunc) {
	l.logAt(hqgologgerlevels.LevelFatal, message, opts...)
}

// Print logs at [LevelSilent]. See [Logger.Log] for filtering and formatting.
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func (l *Logger) Print(message string, opts ...OptionFunc) {
	l.logAt(hqgologgerlevels.LevelSilent, message, opts...)
}

// Error logs at [LevelError]. See [Logger.Log] for filtering and formatting.
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func (l *Logger) Error(message string, opts ...OptionFunc) {
	l.logAt(hqgologgerlevels.LevelError, message, opts...)
}

// Info logs at [LevelInfo]. See [Logger.Log] for filtering and formatting.
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func (l *Logger) Info(message string, opts ...OptionFunc) {
	l.logAt(hqgologgerlevels.LevelInfo, message, opts...)
}

// Warn logs at [LevelWarn]. See [Logger.Log] for filtering and formatting.
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func (l *Logger) Warn(message string, opts ...OptionFunc) {
	l.logAt(hqgologgerlevels.LevelWarn, message, opts...)
}

// Debug logs at [LevelDebug]. See [Logger.Log] for filtering and formatting.
//
// Parameters:
//   - message (string): log message.
//   - opts (...OptionFunc): per-event options.
func (l *Logger) Debug(message string, opts ...OptionFunc) {
	l.logAt(hqgologgerlevels.LevelDebug, message, opts...)
}

// logAt builds an event at the given severity, applies the options, and logs
// it. The severity is re-applied after the options so that neither [WithLevel]
// nor a custom OptionFunc can change it: a Fatal call always exits, and an Info
// call never can.
//
// Parameters:
//   - level (hqgologgerlevels.Level): the event's fixed severity.
//   - message (string): the log message.
//   - opts (...OptionFunc): options applied to the event.
func (l *Logger) logAt(level hqgologgerlevels.Level, message string, opts ...OptionFunc) {
	event := &Event{
		timestamp: time.Now(),
		level:     level,
		message:   message,
		levelSet:  true,
	}

	for _, f := range opts {
		if f != nil {
			f(event)
		}
	}

	event.SetLevel(level)

	l.Log(event)
}

// Log processes a log event by filtering, formatting, and writing it. A nil event is a
// no-op, as is an event whose level was never set (neither [NewEvent] nor
// [Event.SetLevel] ran): such an event is dropped rather than interpreted as
// [LevelFatal]. The event is ignored if its level is greater than the logger's
// threshold (less severe); the threshold is read from a lock-free atomic mirror, so
// filter-discarded events — the common case in production — never touch the shared
// mutex. A concurrent [Logger.SetLevel] may therefore just miss an in-flight event,
// which is then still written; that benign race is inherent to lock-free threshold
// filtering. The event's metadata is passed to the formatter unchanged; renderers
// apply their own conventions — the bundled console formatter substitutes a default
// label based on the level (e.g., "INF" for LevelInfo) when the event carries none.
// Formatting and writing happen only when both a formatter and a writer are
// configured; otherwise the event is silently dropped. Format and write errors are
// deliberately ignored — there is no meaningful recovery path inside a logger, and
// logging must never crash the application. For LevelFatal events the program exits
// with status code 1 regardless of whether the event was written, guaranteeing that
// Fatal never returns; the exit is performed with os.Exit, so deferred functions do
// not run. The method is thread-safe: it holds the logger's read lock around the
// Format and Write calls, so a concurrent [Logger.Close] cannot close the writer in
// the middle of a write. Concurrent Log calls still proceed in parallel with each
// other, so the formatter and writer must be safe for concurrent use. Formatters,
// writers, and colorizers must not call back into the Logger (for example SetLevel,
// SetWriter, Close, or Log on the same logger): Log holds the read lock across those
// calls, so a callback that needs the write lock deadlocks.
//
// Parameters:
//   - event (*Event): the log event to process, containing timestamp, level, message,
//     and metadata.
func (l *Logger) Log(event *Event) {
	if event == nil || !event.levelSet {
		return
	}

	// Lock-free threshold check: events below the configured level are
	// discarded without paying for the shared RWMutex reader count.
	if event.level > hqgologgerlevels.Level(l.threshold.Load()) {
		return
	}

	l.mutex.RLock()
	defer l.mutex.RUnlock()

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
// against the exported [Event] methods. A nil OptionFunc is ignored.
//
// Parameters:
//   - event (*Event): the log event to configure.
type OptionFunc func(event *Event)

// NewEvent creates a new log event at [LevelInfo] with the current time and applies
// the provided OptionFunc configurations to set the level, message, and metadata.
// The metadata map is allocated lazily, only when an option actually attaches
// metadata. Pass [WithLevel] to select another severity; [WithLevel] with
// [LevelFatal] makes the event exit the process when logged. A bare zero-value
// Event is not logged until [Event.SetLevel] is called.
//
// Parameters:
//   - opts (...OptionFunc): Configurations for the log event (e.g., level, message, metadata).
//
// Returns:
//   - event (*Event): a pointer to the configured log event.
func NewEvent(opts ...OptionFunc) (event *Event) {
	event = &Event{
		timestamp: time.Now(),
		level:     hqgologgerlevels.LevelInfo,
		levelSet:  true,
	}

	for _, f := range opts {
		if f != nil {
			f(event)
		}
	}

	return
}

// WithLevel returns an OptionFunc that sets the severity level of a log event. It is
// primarily useful together with [NewEvent] and [Logger.Log] when emitting events
// directly; the level-specific methods (e.g., Info, Fatal) set the level themselves
// after applying options, so they cannot be overridden this way.
//
// Parameters:
//   - level (hqgologgerlevels.Level): the severity level to set.
//
// Returns:
//   - fn (OptionFunc): a function to configure the event's level.
func WithLevel(level hqgologgerlevels.Level) (fn OptionFunc) {
	return func(event *Event) {
		event.SetLevel(level)
	}
}

// WithMessage returns an OptionFunc that sets the message content of a log event.
// It is primarily useful together with [NewEvent] and [Logger.Log] when emitting
// events directly. Options run after the method arguments, so passing it to a
// level-specific method overrides that method's message; the level itself is
// forced after options and cannot be overridden.
//
// Parameters:
//   - message (string): the log message to set.
//
// Returns:
//   - fn (OptionFunc): a function to configure the event's message.
func WithMessage(message string) (fn OptionFunc) {
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
//   - fn (OptionFunc): a function that resets the event's timestamp to the zero value.
func WithoutTimestamp() (fn OptionFunc) {
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
//   - key (string): the metadata key.
//   - value (any): the metadata value, of any type.
//
// Returns:
//   - fn (OptionFunc): a function to configure the event's metadata with the value.
func WithValue(key string, value any) (fn OptionFunc) {
	return func(event *Event) {
		event.SetValue(key, value)
	}
}

// WithString returns an OptionFunc that adds a key-value pair with a string value to a
// log event's metadata. It can be passed to level-specific logging methods (e.g., Info,
// Error) to include custom metadata in the log event.
//
// Parameters:
//   - key (string): the metadata key.
//   - value (string): the metadata value.
//
// Returns:
//   - fn (OptionFunc): a function to configure the event's metadata with a string value.
func WithString(key, value string) (fn OptionFunc) {
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
//   - label (string): the label to set in the metadata.
//
// Returns:
//   - fn (OptionFunc): a function to configure the event's label.
func WithLabel(label string) (fn OptionFunc) {
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
//   - fn (OptionFunc): a function that clears the event's label.
func WithoutLabel() (fn OptionFunc) {
	return func(event *Event) {
		event.SetLabel("")
	}
}

// WithError returns an OptionFunc that adds an error to a log event's metadata
// under the reserved error key. The error is stored as-is; formatters decide how to
// render it. The bundled console formatter prints it as a trailing block
// containing err.Error(), separated from the message by a blank line; a typed-nil
// error (an interface holding a nil pointer) renders as "<nil>" instead of
// panicking. It can be passed to level-specific logging methods to include error
// details.
//
// Parameters:
//   - err (error): the error to set in the metadata.
//
// Returns:
//   - fn (OptionFunc): a function to configure the event's error metadata.
func WithError(err error) (fn OptionFunc) {
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
//   - logger (*Logger): a pointer to a new Logger instance.
func NewLogger() (logger *Logger) {
	logger = &Logger{}

	return
}
