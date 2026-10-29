package logger

import (
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hueristiq/hq-lib-logger-go/formatter"
	"github.com/hueristiq/hq-lib-logger-go/internal/nilvalue"
	"github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/hueristiq/hq-lib-logger-go/writer"
)

// Level is the logger severity type, aliased from the levels package.
type Level = levels.Level

const (
	// LevelFatal is the most severe level; logging at it exits the process.
	LevelFatal = levels.LevelFatal
	// LevelError reports errors that do not halt the program.
	LevelError = levels.LevelError
	// LevelWarn reports conditions that may need attention.
	LevelWarn = levels.LevelWarn
	// LevelInfo reports normal operation.
	LevelInfo = levels.LevelInfo
	// LevelDebug reports diagnostic detail.
	LevelDebug = levels.LevelDebug
	// LevelTrace reports the most verbose detail.
	LevelTrace = levels.LevelTrace
)

// Field is a typed log field, aliased from the formatter package.
type Field = formatter.Field

// String returns a string field.
//
// Parameters:
//   - key (string): field name.
//   - value (string): field value.
//
// Returns:
//   - field (Field): a string field.
func String(key, value string) (field Field) { return formatter.String(key, value) }

// Int returns an integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (int): field value.
//
// Returns:
//   - field (Field): an integer field.
func Int(key string, value int) (field Field) { return formatter.Int(key, value) }

// Int64 returns a signed 64-bit integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (int64): field value.
//
// Returns:
//   - field (Field): an integer field.
func Int64(key string, value int64) (field Field) { return formatter.Int64(key, value) }

// Uint returns an unsigned integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (uint): field value.
//
// Returns:
//   - field (Field): an unsigned integer field.
func Uint(key string, value uint) (field Field) { return formatter.Uint(key, value) }

// Uint64 returns an unsigned 64-bit integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (uint64): field value.
//
// Returns:
//   - field (Field): an unsigned integer field.
func Uint64(key string, value uint64) (field Field) { return formatter.Uint64(key, value) }

// Float64 returns a floating-point field.
//
// Parameters:
//   - key (string): field name.
//   - value (float64): field value.
//
// Returns:
//   - field (Field): a floating-point field.
func Float64(key string, value float64) (field Field) { return formatter.Float64(key, value) }

// Bool returns a boolean field.
//
// Parameters:
//   - key (string): field name.
//   - value (bool): field value.
//
// Returns:
//   - field (Field): a boolean field.
func Bool(key string, value bool) (field Field) { return formatter.Bool(key, value) }

// Duration returns a duration field.
//
// Parameters:
//   - key (string): field name.
//   - value (time.Duration): field value.
//
// Returns:
//   - field (Field): a duration field.
func Duration(key string, value time.Duration) (field Field) { return formatter.Duration(key, value) }

// Time returns a time field.
//
// Parameters:
//   - key (string): field name.
//   - value (time.Time): field value.
//
// Returns:
//   - field (Field): a time field.
func Time(key string, value time.Time) (field Field) { return formatter.Time(key, value) }

// Err returns an error field.
//
// Parameters:
//   - key (string): field name.
//   - value (error): field value.
//
// Returns:
//   - field (Field): an error field.
func Err(key string, value error) (field Field) { return formatter.Err(key, value) }

// Any returns an arbitrary-value field.
//
// Parameters:
//   - key (string): field name.
//   - value (any): field value.
//
// Returns:
//   - field (Field): an arbitrary-value field.
func Any(key string, value any) (field Field) { return formatter.Any(key, value) }

var bufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, 0, 512)

		return &buffer
	},
}

// loggerCore is the configuration shared by a Logger and its child views.
type loggerCore struct {
	mutex        sync.RWMutex
	threshold    atomic.Int64
	reportCaller atomic.Bool
	formatter    formatter.Formatter
	writer       writer.Writer
	hooks        []Hook
	sampler      Sampler
	closed       bool
	closeErr     error
}

type config struct {
	level     levels.Level
	formatter formatter.Formatter
	writer    writer.Writer
	hooks     []Hook
	sampler   Sampler
	caller    bool
}

func defaultConfig() config {
	return config{
		level:     levels.LevelInfo,
		formatter: formatter.NewConsole(nil),
		writer:    writer.NewStream(os.Stderr),
	}
}

// Logger filters, formats, and writes log records. The zero value is usable
// with the same defaults as [New]. A Logger must not be copied after first use.
//
// Fields:
//   - core (atomic.Pointer[loggerCore]): shared configuration core.
//   - fields ([]Field): fields attached to every record from this view.
//   - name (string): dotted child-logger name.
//   - prefix (string): raw text prepended to every message.
type Logger struct {
	core   atomic.Pointer[loggerCore]
	fields []Field
	name   string
	prefix string
}

//nolint:gocritic // config is passed by value to keep callers simple.
func newCore(cfg config) *loggerCore {
	core := &loggerCore{
		formatter: cfg.formatter,
		writer:    cfg.writer,
		hooks:     cfg.hooks,
		sampler:   cfg.sampler,
	}

	core.threshold.Store(int64(cfg.level))
	core.reportCaller.Store(cfg.caller)

	return core
}

func defaultCore() *loggerCore {
	return newCore(defaultConfig())
}

// New returns a logger at [LevelInfo] with a console formatter and a stream
// writer to stderr. Options replace those defaults.
//
// Parameters:
//   - opts (...Option): logger configuration; nil options are ignored.
//
// Returns:
//   - logger (*Logger): the new logger.
func New(opts ...Option) (logger *Logger) {
	cfg := defaultConfig()

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	l := &Logger{}
	l.core.Store(newCore(cfg))

	return l
}

// Option configures a Logger created by [New]. Options are produced by the
// [WithLevel], [WithFormatter], [WithWriter], [WithSampler], [WithHooks], and
// [WithCallerCapture] helpers; nil options are ignored.
type Option func(*config)

// WithLevel sets the severity threshold. An invalid level is ignored.
//
// Parameters:
//   - level (Level): threshold to set.
//
// Returns:
//   - opt (Option): the configuration option.
func WithLevel(level Level) (opt Option) {
	return func(cfg *config) {
		if level.Valid() {
			cfg.level = level
		}
	}
}

// WithFormatter sets the formatter. A nil formatter is ignored.
//
// Parameters:
//   - f (formatter.Formatter): formatter to use.
//
// Returns:
//   - opt (Option): the configuration option.
func WithFormatter(f formatter.Formatter) (opt Option) {
	return func(cfg *config) {
		if !nilvalue.IsNil(f) {
			cfg.formatter = f
		}
	}
}

// WithWriter sets the writer. A nil writer is ignored.
//
// Parameters:
//   - w (writer.Writer): destination writer.
//
// Returns:
//   - opt (Option): the configuration option.
func WithWriter(w writer.Writer) (opt Option) {
	return func(cfg *config) {
		if !nilvalue.IsNil(w) {
			cfg.writer = w
		}
	}
}

// WithSampler sets the sampling policy. A nil sampler disables sampling.
//
// Parameters:
//   - sampler (Sampler): sampling policy.
//
// Returns:
//   - opt (Option): the configuration option.
func WithSampler(sampler Sampler) (opt Option) {
	return func(cfg *config) {
		if nilvalue.IsNil(sampler) {
			cfg.sampler = nil

			return
		}

		cfg.sampler = sampler
	}
}

// WithHooks adds hooks. Nil hooks are ignored.
//
// Parameters:
//   - hooks (...Hook): pre-format hooks.
//
// Returns:
//   - opt (Option): the configuration option.
func WithHooks(hooks ...Hook) (opt Option) {
	return func(cfg *config) {
		for _, hook := range hooks {
			if !nilvalue.IsNil(hook) {
				cfg.hooks = append(cfg.hooks, hook)
			}
		}
	}
}

// WithCallerCapture enables or disables caller capture.
//
// Parameters:
//   - enabled (bool): whether to capture the caller file and line.
//
// Returns:
//   - opt (Option): the configuration option.
func WithCallerCapture(enabled bool) (opt Option) {
	return func(cfg *config) {
		cfg.caller = enabled
	}
}

// getCore returns the shared core, initializing a zero-value Logger lazily.
func (l *Logger) getCore() (core *loggerCore) {
	core = l.core.Load()
	if core != nil {
		return core
	}

	candidate := defaultCore()

	if l.core.CompareAndSwap(nil, candidate) {
		return candidate
	}

	return l.core.Load()
}

// view returns a child sharing this view's core, fields, name, and prefix.
func (l *Logger) view() (child *Logger) {
	child = &Logger{
		fields: l.fields,
		name:   l.name,
		prefix: l.prefix,
	}
	child.core.Store(l.getCore())

	return child
}

// Info logs at [LevelInfo].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Info(msg string, fields ...Field) {
	l.logAt(levels.LevelInfo, msg, fields)
}

// Warn logs at [LevelWarn].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.logAt(levels.LevelWarn, msg, fields)
}

// Error logs at [LevelError].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Error(msg string, fields ...Field) {
	l.logAt(levels.LevelError, msg, fields)
}

// Debug logs at [LevelDebug].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.logAt(levels.LevelDebug, msg, fields)
}

// Trace logs at [LevelTrace].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Trace(msg string, fields ...Field) {
	l.logAt(levels.LevelTrace, msg, fields)
}

// Fatal logs at [LevelFatal] and exits the process with status 1. The exit
// happens even when the record was dropped (for example when the logger is
// closed), and deferred functions do not run.
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.logAt(levels.LevelFatal, msg, fields)
}

// Log logs at an explicit level.
//
// Parameters:
//   - level (Level): record severity.
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Log(level Level, msg string, fields ...Field) {
	l.logAt(level, msg, fields)
}

// LogRecord emits an advanced record with explicit time, caller, and name.
// A zero Time omits the timestamp in formatters that support that.
//
// Parameters:
//   - log (formatter.Log): full record to emit.
//
//nolint:gocritic // Log is passed by value to match the formatter interface.
func (l *Logger) LogRecord(log formatter.Log) {
	core := l.getCore()

	if !log.Level.Valid() || log.Level > Level(core.threshold.Load()) {
		return
	}

	if l.prefix != "" {
		log.Message = l.prefix + log.Message
	}

	if log.Name == "" {
		log.Name = l.name
	}

	log.Fields = mergeFields(l.fields, log.Fields)

	if log.Caller == "" && core.reportCaller.Load() {
		log.Caller = captureCaller()
	}

	l.emit(core, log)
}

// Enabled reports whether level passes the current threshold.
//
// Parameters:
//   - level (Level): severity to test.
//
// Returns:
//   - enabled (bool): true when the level would be emitted.
func (l *Logger) Enabled(level Level) (enabled bool) {
	return level.Valid() && level <= Level(l.getCore().threshold.Load())
}

// Level returns the current threshold.
//
// Returns:
//   - level (Level): the current threshold.
func (l *Logger) Level() (level Level) {
	return Level(l.getCore().threshold.Load())
}

// SetLevel sets the threshold. It returns [levels.ErrUnknownLevel] for an
// invalid level and leaves the threshold unchanged.
//
// Parameters:
//   - level (Level): new threshold.
//
// Returns:
//   - err (error): non-nil when level is invalid.
func (l *Logger) SetLevel(level Level) (err error) {
	if !level.Valid() {
		return levels.ErrUnknownLevel
	}

	l.getCore().threshold.Store(int64(level))

	return nil
}

// Named returns a child view whose records carry name.
//
// Parameters:
//   - name (string): child name; empty returns an equivalent child.
//
// Returns:
//   - child (*Logger): the named child view.
func (l *Logger) Named(name string) (child *Logger) {
	child = l.view()

	if name == "" {
		return child
	}

	if child.name == "" {
		child.name = name
	} else {
		child.name += "." + name
	}

	return child
}

// With returns a child view with fields attached.
//
// Parameters:
//   - fields (...Field): fields to attach.
//
// Returns:
//   - child (*Logger): the child view.
func (l *Logger) With(fields ...Field) (child *Logger) {
	child = l.view()

	if len(fields) == 0 {
		return child
	}

	child.fields = mergeFields(l.fields, fields)

	return child
}

// WithPrefix returns a child view whose messages are prefixed with prefix.
//
// Parameters:
//   - prefix (string): raw message prefix.
//
// Returns:
//   - child (*Logger): the prefixed child view.
func (l *Logger) WithPrefix(prefix string) (child *Logger) {
	child = l.view()

	if prefix != "" {
		child.prefix += prefix
	}

	return child
}

// AsWriter returns an io.Writer that logs each write at [LevelInfo].
//
// Returns:
//   - out (io.Writer): the adapter.
func (l *Logger) AsWriter() (out io.Writer) {
	return logWriter{logger: l}
}

// Sync flushes the writer when it supports Sync or Flush.
//
// Returns:
//   - err (error): the writer's flush error, if any.
func (l *Logger) Sync() (err error) {
	core := l.getCore()

	core.mutex.RLock()
	defer core.mutex.RUnlock()

	if core.writer == nil {
		return nil
	}

	if syncer, ok := core.writer.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}

	if flusher, ok := core.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}

	return nil
}

// Close closes the writer. Repeated calls return the first error.
//
// Returns:
//   - err (error): the writer's close error, if any.
func (l *Logger) Close() (err error) {
	core := l.getCore()

	core.mutex.Lock()
	defer core.mutex.Unlock()

	if core.closed {
		return core.closeErr
	}

	core.closed = true

	if core.writer == nil {
		return nil
	}

	core.closeErr = core.writer.Close()

	return core.closeErr
}

func (l *Logger) logAt(level Level, msg string, fields []Field) {
	if !level.Valid() {
		return
	}

	l.LogRecord(formatter.Log{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		Fields:  fields,
	})
}

// emit formats and writes one record. Formatter and writer errors are
// deliberately discarded: a logger cannot report its own failure.
//
//nolint:gocritic // Log is passed by value to match the Appender interface.
func (l *Logger) emit(core *loggerCore, log formatter.Log) {
	fatal := log.Level == LevelFatal

	core.mutex.RLock()

	func() {
		defer core.mutex.RUnlock()

		if core.closed || core.formatter == nil || core.writer == nil {
			return
		}

		if !fatal && core.sampler != nil && !core.sampler.Sample(log.Level) {
			return
		}

		for _, hook := range core.hooks {
			_ = hook.Fire(&log)
		}

		if appender, ok := core.formatter.(formatter.Appender); ok {
			buffer := bufferPool.Get().(*[]byte)

			data, err := appender.Append((*buffer)[:0], log)
			if err == nil {
				_ = core.writer.Write(data, log.Level)
			}

			*buffer = data[:0]
			bufferPool.Put(buffer)

			return
		}

		if data, err := core.formatter.Format(&log); err == nil {
			_ = core.writer.Write(data, log.Level)
		}
	}()

	if fatal {
		os.Exit(1)
	}
}

func mergeFields(base, extra []Field) []Field {
	if len(base) == 0 {
		return extra
	}

	if len(extra) == 0 {
		return base
	}

	combined := make([]Field, 0, len(base)+len(extra))
	combined = append(combined, base...)
	combined = append(combined, extra...)

	return combined
}

type logWriter struct {
	logger *Logger
}

func (w logWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(strings.TrimSuffix(string(p), "\n"))

	return len(p), nil
}
