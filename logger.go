package logger

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	"github.com/hueristiq/hq-lib-logger-go/internal/nilvalue"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

// Level is the logger severity type, aliased from the levels package.
type Level = hqgologgerlevels.Level

const (
	// LevelFatal is the most severe level; logging at it exits the process.
	LevelFatal = hqgologgerlevels.LevelFatal
	// LevelError reports errors that do not halt the program.
	LevelError = hqgologgerlevels.LevelError
	// LevelWarn reports conditions that may need attention.
	LevelWarn = hqgologgerlevels.LevelWarn
	// LevelInfo reports normal operation.
	LevelInfo = hqgologgerlevels.LevelInfo
	// LevelDebug reports diagnostic detail.
	LevelDebug = hqgologgerlevels.LevelDebug
	// LevelTrace reports the most verbose detail.
	LevelTrace = hqgologgerlevels.LevelTrace
)

// Field is a typed log field, aliased from the formatter package.
type Field = hqgologgerformatter.Field

// String returns a string field.
//
// Parameters:
//   - key (string): field name.
//   - value (string): field value.
//
// Returns:
//   - field (Field): a string field.
func String(key, value string) (field Field) { return hqgologgerformatter.String(key, value) }

// Int returns an integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (int): field value.
//
// Returns:
//   - field (Field): an integer field.
func Int(key string, value int) (field Field) { return hqgologgerformatter.Int(key, value) }

// Int64 returns a signed 64-bit integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (int64): field value.
//
// Returns:
//   - field (Field): an integer field.
func Int64(key string, value int64) (field Field) { return hqgologgerformatter.Int64(key, value) }

// Uint returns an unsigned integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (uint): field value.
//
// Returns:
//   - field (Field): an unsigned integer field.
func Uint(key string, value uint) (field Field) { return hqgologgerformatter.Uint(key, value) }

// Uint64 returns an unsigned 64-bit integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (uint64): field value.
//
// Returns:
//   - field (Field): an unsigned integer field.
func Uint64(key string, value uint64) (field Field) { return hqgologgerformatter.Uint64(key, value) }

// Float64 returns a floating-point field.
//
// Parameters:
//   - key (string): field name.
//   - value (float64): field value.
//
// Returns:
//   - field (Field): a floating-point field.
func Float64(key string, value float64) (field Field) { return hqgologgerformatter.Float64(key, value) }

// Bool returns a boolean field.
//
// Parameters:
//   - key (string): field name.
//   - value (bool): field value.
//
// Returns:
//   - field (Field): a boolean field.
func Bool(key string, value bool) (field Field) { return hqgologgerformatter.Bool(key, value) }

// Duration returns a duration field.
//
// Parameters:
//   - key (string): field name.
//   - value (time.Duration): field value.
//
// Returns:
//   - field (Field): a duration field.
func Duration(key string, value time.Duration) (field Field) {
	return hqgologgerformatter.Duration(key, value)
}

// Time returns a time field.
//
// Parameters:
//   - key (string): field name.
//   - value (time.Time): field value.
//
// Returns:
//   - field (Field): a time field.
func Time(key string, value time.Time) (field Field) { return hqgologgerformatter.Time(key, value) }

// Err returns an error field.
//
// Parameters:
//   - key (string): field name.
//   - value (error): field value.
//
// Returns:
//   - field (Field): an error field.
func Err(key string, value error) (field Field) { return hqgologgerformatter.Err(key, value) }

// Any returns an arbitrary-value field.
//
// Parameters:
//   - key (string): field name.
//   - value (any): field value.
//
// Returns:
//   - field (Field): an arbitrary-value field.
func Any(key string, value any) (field Field) { return hqgologgerformatter.Any(key, value) }

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
	formatter    hqgologgerformatter.Formatter
	writer       hqgologgerwriter.Writer
	hooks        []Hook
	sampler      Sampler
	closed       bool
	closeErr     error
}

// config is the builder state consumed by [newCore] when constructing a
// logger.
//
// Fields:
//   - level (hqgologgerlevels.Level): severity threshold.
//   - formatter (hqgologgerformatter.Formatter): record formatter.
//   - writer (hqgologgerwriter.Writer): destination writer.
//   - hooks ([]Hook): pre-format hooks.
//   - sampler (Sampler): sampling policy.
//   - caller (bool): whether to capture callers.
type config struct {
	level     hqgologgerlevels.Level
	formatter hqgologgerformatter.Formatter
	writer    hqgologgerwriter.Writer
	hooks     []Hook
	sampler   Sampler
	caller    bool
}

// defaultConfig returns the package defaults: LevelInfo, a console
// formatter, and a stream writer on stderr.
//
// Returns:
//   - cfg (config): the default configuration.
func defaultConfig() config {
	return config{
		level:     hqgologgerlevels.LevelInfo,
		formatter: hqgologgerformatter.NewConsole(nil),
		writer:    hqgologgerwriter.NewStream(os.Stderr),
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

// newCore builds a core from the given configuration.
//
// Parameters:
//   - cfg (config): configuration to apply.
//
// Returns:
//   - core (*loggerCore): the initialized core.
//
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

// defaultCore builds a core from the package defaults.
//
// Returns:
//   - core (*loggerCore): the default core.
func defaultCore() *loggerCore {
	return newCore(defaultConfig())
}

// New returns a logger at [LevelInfo] with a console formatter and a stream
// writer to stderr. Options replace those defaults.
//
// Parameters:
//   - opts (...OptionFunc): logger configuration; nil options are ignored.
//
// Returns:
//   - logger (*Logger): the new logger.
func New(opts ...OptionFunc) (logger *Logger) {
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

// OptionFunc configures a Logger created by [New]. Options are produced by the
// [WithLevel], [WithFormatter], [WithWriter], [WithSampler], [WithHooks], and
// [WithCallerCapture] helpers; nil options are ignored.
type OptionFunc func(*config)

// WithLevel sets the severity threshold. An invalid level is ignored.
//
// Parameters:
//   - level (Level): threshold to set.
//
// Returns:
//   - opt (OptionFunc): the configuration option.
func WithLevel(level Level) (opt OptionFunc) {
	return func(cfg *config) {
		if level.Valid() {
			cfg.level = level
		}
	}
}

// WithFormatter sets the formatter. A nil formatter is ignored.
//
// Parameters:
//   - f (hqgologgerformatter.Formatter): formatter to use.
//
// Returns:
//   - opt (OptionFunc): the configuration option.
func WithFormatter(f hqgologgerformatter.Formatter) (opt OptionFunc) {
	return func(cfg *config) {
		if !nilvalue.IsNil(f) {
			cfg.formatter = f
		}
	}
}

// WithWriter sets the writer. A nil writer is ignored.
//
// Parameters:
//   - w (hqgologgerwriter.Writer): destination writer.
//
// Returns:
//   - opt (OptionFunc): the configuration option.
func WithWriter(w hqgologgerwriter.Writer) (opt OptionFunc) {
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
//   - opt (OptionFunc): the configuration option.
func WithSampler(sampler Sampler) (opt OptionFunc) {
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
//   - opt (OptionFunc): the configuration option.
func WithHooks(hooks ...Hook) (opt OptionFunc) {
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
//   - opt (OptionFunc): the configuration option.
func WithCallerCapture(enabled bool) (opt OptionFunc) {
	return func(cfg *config) {
		cfg.caller = enabled
	}
}

// getCore returns the shared core, initializing a zero-value Logger lazily.
//
// Returns:
//   - core (*loggerCore): the shared core.
func (l *Logger) getCore() (core *loggerCore) {
	core = l.core.Load()
	if core != nil {
		return core
	}

	candidate := defaultCore()

	if l.core.CompareAndSwap(nil, candidate) {
		return candidate
	}

	// A concurrent first use may lose the CAS: another goroutine already
	// stored its core, so load and return that one.
	return l.core.Load()
}

// view returns a child sharing this view's core, fields, name, and prefix.
//
// Returns:
//   - child (*Logger): the child view.
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
	l.logAt(hqgologgerlevels.LevelInfo, msg, fields)
}

// Infof logs a formatted message at [LevelInfo]. The formatting allocates;
// prefer [Logger.Info] with typed fields on hot paths.
//
// Parameters:
//   - format (string): printf-style format string.
//   - args (...any): format arguments.
func (l *Logger) Infof(format string, args ...any) {
	l.logAt(hqgologgerlevels.LevelInfo, fmt.Sprintf(format, args...), nil)
}

// Warn logs at [LevelWarn].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.logAt(hqgologgerlevels.LevelWarn, msg, fields)
}

// Warnf logs a formatted message at [LevelWarn]. The formatting allocates;
// prefer [Logger.Warn] with typed fields on hot paths.
//
// Parameters:
//   - format (string): printf-style format string.
//   - args (...any): format arguments.
func (l *Logger) Warnf(format string, args ...any) {
	l.logAt(hqgologgerlevels.LevelWarn, fmt.Sprintf(format, args...), nil)
}

// Error logs at [LevelError].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Error(msg string, fields ...Field) {
	l.logAt(hqgologgerlevels.LevelError, msg, fields)
}

// Errorf logs a formatted message at [LevelError]. The formatting allocates;
// prefer [Logger.Error] with typed fields on hot paths.
//
// Parameters:
//   - format (string): printf-style format string.
//   - args (...any): format arguments.
func (l *Logger) Errorf(format string, args ...any) {
	l.logAt(hqgologgerlevels.LevelError, fmt.Sprintf(format, args...), nil)
}

// Debug logs at [LevelDebug].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.logAt(hqgologgerlevels.LevelDebug, msg, fields)
}

// Debugf logs a formatted message at [LevelDebug]. The formatting allocates;
// prefer [Logger.Debug] with typed fields on hot paths.
//
// Parameters:
//   - format (string): printf-style format string.
//   - args (...any): format arguments.
func (l *Logger) Debugf(format string, args ...any) {
	l.logAt(hqgologgerlevels.LevelDebug, fmt.Sprintf(format, args...), nil)
}

// Trace logs at [LevelTrace].
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Trace(msg string, fields ...Field) {
	l.logAt(hqgologgerlevels.LevelTrace, msg, fields)
}

// Tracef logs a formatted message at [LevelTrace]. The formatting allocates;
// prefer [Logger.Trace] with typed fields on hot paths.
//
// Parameters:
//   - format (string): printf-style format string.
//   - args (...any): format arguments.
func (l *Logger) Tracef(format string, args ...any) {
	l.logAt(hqgologgerlevels.LevelTrace, fmt.Sprintf(format, args...), nil)
}

// Fatal logs at [LevelFatal] and exits the process with status 1. The exit
// happens even when the record was dropped (for example when the logger is
// closed), and deferred functions do not run.
//
// Parameters:
//   - msg (string): log message.
//   - fields (...Field): typed context.
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.logAt(hqgologgerlevels.LevelFatal, msg, fields)
}

// Fatalf logs a formatted message at [LevelFatal] and exits the process with
// status 1, exactly like [Logger.Fatal].
//
// Parameters:
//   - format (string): printf-style format string.
//   - args (...any): format arguments.
func (l *Logger) Fatalf(format string, args ...any) {
	l.logAt(hqgologgerlevels.LevelFatal, fmt.Sprintf(format, args...), nil)
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

// Logf logs a formatted message at an explicit level. The formatting
// allocates; prefer [Logger.Log] with typed fields on hot paths.
//
// Parameters:
//   - level (Level): record severity.
//   - format (string): printf-style format string.
//   - args (...any): format arguments.
func (l *Logger) Logf(level Level, format string, args ...any) {
	l.logAt(level, fmt.Sprintf(format, args...), nil)
}

// LogRecord emits an advanced record with explicit time, caller, and name.
// A zero Time omits the timestamp in formatters that support that.
//
// Parameters:
//   - log (hqgologgerformatter.Log): full record to emit.
//
//nolint:gocritic // Log is passed by value to match the formatter interface.
func (l *Logger) LogRecord(log hqgologgerformatter.Log) {
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

// SetLevel sets the threshold. It returns [hqgologgerlevels.ErrUnknownLevel] for an
// invalid level and leaves the threshold unchanged.
//
// Parameters:
//   - level (Level): new threshold.
//
// Returns:
//   - err (error): non-nil when level is invalid.
func (l *Logger) SetLevel(level Level) (err error) {
	if !level.Valid() {
		return hqgologgerlevels.ErrUnknownLevel
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

// logAt emits a record at the given level with the current time. Invalid
// levels are dropped.
//
// Parameters:
//   - level (Level): record severity.
//   - msg (string): log message.
//   - fields ([]Field): typed context.
func (l *Logger) logAt(level Level, msg string, fields []Field) {
	if !level.Valid() {
		return
	}

	l.LogRecord(hqgologgerformatter.Log{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		Fields:  fields,
	})
}

// emit formats and writes one record. Formatter and writer errors are
// deliberately discarded: a logger cannot report its own failure.
//
// Parameters:
//   - core (*loggerCore): the configuration core.
//   - log (hqgologgerformatter.Log): record to emit.
//
//nolint:gocritic // Log is passed by value to match the Appender interface.
func (l *Logger) emit(core *loggerCore, log hqgologgerformatter.Log) {
	fatal := log.Level == LevelFatal

	core.mutex.RLock()
	defer core.mutex.RUnlock()

	if core.closed || core.formatter == nil || core.writer == nil {
		exitOnFatal(fatal)

		return
	}

	if !fatal && core.sampler != nil && !core.sampler.Sample(log.Level) {
		exitOnFatal(fatal)

		return
	}

	for _, hook := range core.hooks {
		if filter, ok := hook.(LevelFilter); ok && !slices.Contains(filter.Levels(), log.Level) {
			continue
		}

		_ = hook.Fire(log)
	}

	if appender, ok := core.formatter.(hqgologgerformatter.Appender); ok {
		buffer := bufferPool.Get().(*[]byte)

		data, err := appender.Append((*buffer)[:0], log)
		if err == nil {
			_ = core.writer.Write(data, log.Level)
		}

		*buffer = data[:0]
		bufferPool.Put(buffer)
	} else if data, err := core.formatter.Format(log); err == nil {
		_ = core.writer.Write(data, log.Level)
	}

	exitOnFatal(fatal)
}

// exitOnFatal exits the process with status 1 when fatal is true.
//
// Parameters:
//   - fatal (bool): whether the record was fatal.
func exitOnFatal(fatal bool) {
	if fatal {
		os.Exit(1)
	}
}

// mergeFields concatenates base and extra field slices. Empty inputs return
// the other slice unchanged, avoiding a copy for the common single-view case.
//
// Parameters:
//   - base ([]Field): logger-view fields.
//   - extra ([]Field): per-record fields.
//
// Returns:
//   - combined ([]Field): the concatenated fields.
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

// logWriter adapts a Logger to io.Writer for [Logger.AsWriter].
//
// Fields:
//   - logger (*Logger): target logger.
type logWriter struct {
	logger *Logger
}

// Write logs each write as a single [LevelInfo] record, trimming one
// trailing newline.
//
// Parameters:
//   - p ([]byte): text to log.
//
// Returns:
//   - n (int): always len(p).
//   - err (error): always nil.
func (w logWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(strings.TrimSuffix(string(p), "\n"))

	return len(p), nil
}
