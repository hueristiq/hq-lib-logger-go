package writer

import (
	"errors"
	"io"
	"sync"

	"github.com/hueristiq/hq-lib-logger-go/internal/nilvalue"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// MultiWriter is an implementation of the Writer interface that aggregates
// multiple Writer instances, forwarding log messages to each underlying writer.
// It enables simultaneous logging to multiple destinations (e.g., console and file)
// while maintaining a single Writer interface. Every writer is always attempted,
// even if some fail; the failures are combined with [errors.Join]. Nil and
// typed-nil writers are filtered out during initialization to prevent runtime
// issues. A writer listed more than once is closed once per entry; idempotent
// child Close makes that harmless. Construct with [NewMultiWriter]. A MultiWriter
// must not be copied after first use.
//
// Fields:
//   - writers ([]Writer): destinations; immutable after construction.
//   - once (sync.Once): ensures Close runs once.
//   - closeErr (error): combined result of the first Close.
type MultiWriter struct {
	writers  []Writer
	once     sync.Once
	closeErr error
}

// Write forwards the provided log data and severity level to each underlying
// Writer in the MultiWriter's writers slice. It attempts to write to all writers,
// even if some fail, and returns all failures combined with [errors.Join]. The
// method is thread-safe as long as the underlying writers are thread-safe.
//
// Parameters:
//   - data ([]byte): the pre-formatted log message to write, typically produced
//     by a formatter.
//   - level (hqgologgerlevels.Level): the severity level of the log message, as defined in
//     the levels package (e.g., LevelFatal, LevelDebug), used by writers to filter
//     messages based on their configured thresholds.
//
// Returns:
//   - err (error): the combined errors from the underlying writers, or nil if
//     all writes succeed or no writers are present.
func (m *MultiWriter) Write(data []byte, level hqgologgerlevels.Level) (err error) {
	var errs []error

	for _, w := range m.writers {
		if werr := w.Write(data, level); werr != nil {
			errs = append(errs, werr)
		}
	}

	return errors.Join(errs...)
}

// Close closes all underlying writers in the MultiWriter's writers slice,
// releasing their associated resources. It attempts to close all writers, even
// if some fail, and combines their errors with [errors.Join]. Repeated calls are
// safe and return the first call's combined error. The method is thread-safe as
// long as the underlying writers are thread-safe.
//
// Returns:
//   - err (error): the combined errors from the underlying writers, or nil if
//     all closes succeed or no writers are present.
func (m *MultiWriter) Close() (err error) {
	m.once.Do(func() {
		var errs []error

		for _, w := range m.writers {
			if cerr := w.Close(); cerr != nil {
				errs = append(errs, cerr)
			}
		}

		m.closeErr = errors.Join(errs...)
	})

	return m.closeErr
}

// Writer defines the interface for writing log messages to an output destination.
// Implementations of this interface handle the delivery of formatted log data to
// specific sinks, such as files, consoles, network endpoints, or external logging
// services, while considering the severity level of the log message. The interface
// extends io.Closer to ensure resources (e.g., file handles or network connections)
// can be properly closed when logging is complete; the Logger that owns a writer
// closes it through the logger's Close method, so implementations without
// resources may simply return nil from Close. Implementations are expected to be
// safe for concurrent use, since a Logger may write from several goroutines at
// once.
//
// Methods:
//   - Write(data []byte, level hqgologgerlevels.Level) (err error): writes the
//     formatted record to the destination.
//   - Close() (err error): releases the writer's resources.
type Writer interface {
	io.Closer

	// Write writes data to the destination.
	//
	// Parameters:
	//   - data ([]byte): the formatted log record.
	//   - level (hqgologgerlevels.Level): the record severity.
	//
	// Returns:
	//   - err (error): non-nil when the write fails.
	Write(data []byte, level hqgologgerlevels.Level) (err error)
}

// Compile-time guard ensuring [MultiWriter] satisfies [Writer].
var _ Writer = (*MultiWriter)(nil)

// NewMultiWriter creates and returns a new MultiWriter instance that aggregates
// the provided Writer instances. It filters out nil and typed-nil writers (an
// interface holding a nil pointer, which compares unequal to nil but panics when
// its methods are called) to ensure safe operation, and initializes the writers
// slice with the remaining writers. The resulting MultiWriter can be used to
// forward log messages to multiple destinations simultaneously. If no usable
// writers are provided, an empty MultiWriter is returned, which performs no
// operations when used.
//
// Parameters:
//   - writers (...Writer): a variadic list of Writer instances to aggregate.
//
// Returns:
//   - multi (*MultiWriter): a pointer to a new MultiWriter instance containing
//     the non-nil writers.
func NewMultiWriter(writers ...Writer) (multi *MultiWriter) {
	multi = &MultiWriter{
		writers: make([]Writer, 0, len(writers)),
	}

	for _, w := range writers {
		if !isNilWriter(w) {
			multi.writers = append(multi.writers, w)
		}
	}

	return
}

// isNilWriter reports whether w is nil, including a typed-nil Writer of a
// nilable kind (pointer, map, slice, func, chan, or interface).
//
// Parameters:
//   - w (Writer): the destination writer.
//
// Returns:
//   - ok (bool): whether the condition holds.
func isNilWriter(w Writer) (ok bool) {
	return nilvalue.IsNil(w)
}
