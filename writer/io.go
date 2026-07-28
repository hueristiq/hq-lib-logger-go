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
//   - mutex (sync.Mutex): Serializes writes to the underlying writer.
//   - w (io.Writer): The destination receiving each message.
type IOWriter struct {
	mutex sync.Mutex
	w     io.Writer
}

// Write appends a newline to data and delivers it to the underlying writer in a
// single write, flushing afterwards when the writer supports it. The severity
// level is accepted to satisfy the [Writer] interface but does not affect
// routing: every level is written. The caller's slice — including its spare
// capacity — is never modified. The method is thread-safe.
//
// Parameters:
//   - data ([]byte): The pre-formatted log message to write, typically produced
//     by a formatter.
//   - level (hqgologgerlevels.Level): The severity level of the log message; ignored.
//
// Returns:
//   - err (error): An error if writing to the destination fails or if flushing
//     a flushable destination fails; nil otherwise.
func (a *IOWriter) Write(data []byte, _ hqgologgerlevels.Level) (err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	line := make([]byte, len(data)+1)
	copy(line, data)
	line[len(data)] = '\n'

	if _, err = a.w.Write(line); err != nil {
		return
	}

	if flusher, ok := a.w.(interface{ Flush() error }); ok {
		err = flusher.Flush()
	}

	return
}

// Close closes the underlying writer if it implements [io.Closer], releasing its
// resources; otherwise it is a no-op. The method is thread-safe.
//
// Returns:
//   - err (error): The error from closing the underlying writer, or nil if the
//     close succeeds or the writer is not closable.
func (a *IOWriter) Close() (err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if closer, ok := a.w.(io.Closer); ok {
		err = closer.Close()
	}

	return
}

var _ Writer = (*IOWriter)(nil)

// NewIOWriter creates and returns a new IOWriter that forwards log messages to w.
// If w is nil, [io.Discard] is used instead, so the adapter is always safe to
// write to.
//
// Parameters:
//   - w (io.Writer): The destination for log messages (e.g., an *os.File).
//
// Returns:
//   - adapter (*IOWriter): A pointer to a new IOWriter wrapping w.
func NewIOWriter(w io.Writer) (adapter *IOWriter) {
	if w == nil {
		w = io.Discard
	}

	adapter = &IOWriter{
		w: w,
	}

	return
}
