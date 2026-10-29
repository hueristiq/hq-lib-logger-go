// Package writer delivers formatted log bytes to output destinations.
//
// A [Writer] writes a preformatted message with its severity and can be closed.
// [Stream] adapts any io.Writer, [MultiWriter] fans out to several writers,
// [Syslog] writes to Unix syslog, and [EventLog] writes to the Windows event
// log. Platform writers return [ErrUnsupported] where their sink is absent.
//
// Writers are called concurrently and must be safe for concurrent use. Write
// must not retain the data slice after it returns.
package writer

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/hueristiq/hq-lib-logger-go/internal/nilvalue"
	"github.com/hueristiq/hq-lib-logger-go/levels"
)

// Writer delivers a formatted record to a destination.
//
// Methods:
//   - Write(data []byte, level levels.Level) (err error): delivers a record.
//   - Close() (err error): releases the writer's resources.
type Writer interface {
	Write(data []byte, level levels.Level) (err error)
	Close() (err error)
}

// MultiWriter forwards each record to several writers.
//
// Fields:
//   - writers ([]Writer): destinations; immutable after construction.
//   - once (sync.Once): ensures Close runs once.
//   - closed (atomic.Bool): whether Close has run.
//   - closeErr (error): combined result of the first Close.
type MultiWriter struct {
	writers  []Writer
	once     sync.Once
	closed   atomic.Bool
	closeErr error
}

// NewMulti returns a MultiWriter that forwards to writers. Nil writers are
// ignored.
//
// Parameters:
//   - writers (...Writer): destinations to fan out to.
//
// Returns:
//   - multi (*MultiWriter): the fan-out writer.
func NewMulti(writers ...Writer) (multi *MultiWriter) {
	multi = &MultiWriter{writers: make([]Writer, 0, len(writers))}

	for _, w := range writers {
		if !nilvalue.IsNil(w) {
			multi.writers = append(multi.writers, w)
		}
	}

	return multi
}

// Write forwards data to every writer and joins their errors.
//
// Parameters:
//   - data ([]byte): formatted record.
//   - level (levels.Level): record severity.
//
// Returns:
//   - err (error): joined errors from the destinations.
func (m *MultiWriter) Write(data []byte, level levels.Level) (err error) {
	if m == nil || m.closed.Load() {
		return ErrClosed
	}

	var errs []error

	for _, w := range m.writers {
		if writeErr := w.Write(data, level); writeErr != nil {
			errs = append(errs, writeErr)
		}
	}

	return errors.Join(errs...)
}

// Sync flushes every writer that supports Sync or Flush and joins their errors.
//
// Returns:
//   - err (error): joined errors from the destinations.
func (m *MultiWriter) Sync() (err error) {
	if m == nil || m.closed.Load() {
		return ErrClosed
	}

	var errs []error

	for _, w := range m.writers {
		if syncer, ok := w.(interface{ Sync() error }); ok {
			if syncErr := syncer.Sync(); syncErr != nil {
				errs = append(errs, syncErr)
			}

			continue
		}

		if flusher, ok := w.(interface{ Flush() error }); ok {
			if flushErr := flusher.Flush(); flushErr != nil {
				errs = append(errs, flushErr)
			}
		}
	}

	return errors.Join(errs...)
}

// Close closes every writer and joins their errors. Repeated calls return the
// first result. Write and Sync return [ErrClosed] afterwards.
//
// Returns:
//   - err (error): joined errors from the destinations.
func (m *MultiWriter) Close() (err error) {
	if m == nil {
		return nil
	}

	m.once.Do(func() {
		m.closed.Store(true)

		var errs []error

		for _, w := range m.writers {
			if closeErr := w.Close(); closeErr != nil {
				errs = append(errs, closeErr)
			}
		}

		m.closeErr = errors.Join(errs...)
	})

	return m.closeErr
}

var _ Writer = (*MultiWriter)(nil)
