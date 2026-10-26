package writer

import (
	"io"
	"sync"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// IOWriter adapts a standard [io.Writer] — a file, a network connection, a
// buffer — to the [Writer] interface, so ordinary destinations can receive log
// output without a hand-written wrapper. Every message is passed through
// regardless of its severity level; filtering is the logger's job. A newline is
// appended to each message (formatters do not emit one), and if the underlying
// writer supports flushing (via a Flush method) it is flushed after each write.
// Writes are serialized with a mutex, making the adapter safe for concurrent
// use even when the underlying writer is not. Construct with [NewIOWriter].
// An IOWriter must not be copied after first use.
//
// Fields:
//   - mutex (sync.Mutex): serializes writes to the underlying writer.
//   - w (io.Writer): destination receiving each message.
//   - buf ([]byte): scratch space for the trailing newline, guarded by mutex.
//   - closed (bool): whether Close has run, guarded by mutex.
//   - closeErr (error): result of the first Close, guarded by mutex.
type IOWriter struct {
	mutex    sync.Mutex
	w        io.Writer
	buf      []byte
	closed   bool
	closeErr error
}

// Write appends a newline to data and delivers it to the underlying writer in a
// single write, flushing afterwards when the writer supports it. The severity
// level is accepted to satisfy the [Writer] interface but does not affect
// routing: every level is written. The caller's slice — including its spare
// capacity — is never modified: the newline is appended in a reusable scratch
// buffer retained between writes (sized to the largest line seen), so steady-state
// writes allocate nothing. The method is thread-safe.
//
// Parameters:
//   - data ([]byte): the pre-formatted log message to write, typically produced
//     by a formatter.
//
// Returns:
//   - err (error): an error if writing to the destination fails or if flushing
//     a flushable destination fails; nil otherwise.
func (a *IOWriter) Write(data []byte, _ hqgologgerlevels.Level) (err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Reuse the scratch buffer instead of allocating a per-line copy. Safe
	// because Write is fully serialized by the mutex and the io.Writer contract
	// forbids the destination from retaining the slice it is given.
	a.buf = append(a.buf[:0], data...)
	a.buf = append(a.buf, '\n')

	if _, err = a.w.Write(a.buf); err != nil {
		return
	}

	if flusher, ok := a.w.(interface{ Flush() error }); ok {
		err = flusher.Flush()
	}

	return
}

// Close closes the underlying writer if it implements [io.Closer], releasing its
// resources; otherwise it is a no-op. Repeated calls are safe and return the first
// call's error. The method is thread-safe.
//
// Returns:
//   - err (error): the error from closing the underlying writer, or nil if the
//     close succeeds or the writer is not closable.
func (a *IOWriter) Close() (err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.closed {
		return a.closeErr
	}

	a.closed = true

	if closer, ok := a.w.(io.Closer); ok {
		a.closeErr = closer.Close()
	}

	return a.closeErr
}

// Compile-time guard ensuring [IOWriter] satisfies [Writer].
var _ Writer = (*IOWriter)(nil)

// NewIOWriter creates and returns a new IOWriter that forwards log messages to w.
// If w is nil, [io.Discard] is used instead, so the adapter is always safe to
// write to. A typed-nil writer is not nil and is used as-is.
//
// Parameters:
//   - w (io.Writer): the destination for log messages (e.g., an *os.File).
//
// Returns:
//   - adapter (*IOWriter): a pointer to a new IOWriter wrapping w.
func NewIOWriter(w io.Writer) (adapter *IOWriter) {
	if w == nil {
		w = io.Discard
	}

	adapter = &IOWriter{
		w: w,
	}

	return
}
