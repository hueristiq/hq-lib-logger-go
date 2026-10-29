package writer

import (
	"io"
	"os"
	"sync"

	"github.com/hueristiq/hq-lib-logger-go/internal/nilvalue"
	"github.com/hueristiq/hq-lib-logger-go/levels"
)

// Stream adapts an io.Writer. Every record is written with one trailing
// newline and flushed when the destination supports Flush. Writes are
// serialized, and a nil destination becomes [io.Discard].
//
// Fields:
//   - mutex (sync.Mutex): serializes writes and protects buffer and state.
//   - writer (io.Writer): destination.
//   - buffer ([]byte): reusable scratch buffer for the newline.
//   - closed (bool): whether Close has run.
//   - closeErr (error): result of the first Close.
type Stream struct {
	mutex    sync.Mutex
	writer   io.Writer
	buffer   []byte
	closed   bool
	closeErr error
}

// NewStream returns a Stream writing to w. A nil writer becomes [io.Discard].
//
// Parameters:
//   - w (io.Writer): destination writer; nil becomes io.Discard.
//
// Returns:
//   - stream (*Stream): the writer.
func NewStream(w io.Writer) (stream *Stream) {
	if nilvalue.IsNil(w) {
		w = io.Discard
	}

	return &Stream{writer: w}
}

// Write appends one newline to data and writes it to the destination.
//
// Parameters:
//   - data ([]byte): formatted record.
//   - level (levels.Level): record severity; ignored.
//
// Returns:
//   - err (error): a write or flush error.
func (s *Stream) Write(data []byte, _ levels.Level) (err error) {
	if s == nil || s.writer == nil {
		return ErrClosed
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return ErrClosed
	}

	s.buffer = append(s.buffer[:0], data...)
	s.buffer = append(s.buffer, '\n')

	written, err := s.writer.Write(s.buffer)
	if err != nil {
		return err
	}

	if written != len(s.buffer) {
		return io.ErrShortWrite
	}

	if flusher, ok := s.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}

	return nil
}

// Close closes the destination when it implements io.Closer. Repeated calls
// return the first result, and Write returns [ErrClosed] afterwards.
//
// Returns:
//   - err (error): the close error, if any.
func (s *Stream) Close() (err error) {
	if s == nil || s.writer == nil {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return s.closeErr
	}

	s.closed = true

	if s.writer == os.Stdout || s.writer == os.Stderr {
		return nil
	}

	if closer, ok := s.writer.(io.Closer); ok {
		s.closeErr = closer.Close()
	}

	return s.closeErr
}

var _ Writer = (*Stream)(nil)
