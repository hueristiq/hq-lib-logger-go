//nolint:testpackage // tests exercise unexported behavior.
package writer

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	return len(p) - 1, nil
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

type flushTracker struct {
	bytes.Buffer

	flushed bool
}

func (f *flushTracker) Flush() error {
	f.flushed = true

	return nil
}

func TestStreamWritesNewline(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	stream := NewStream(&buffer)

	require.NoError(t, stream.Write([]byte("hello"), hqgologgerlevel.LevelInfo))
	assert.Equal(t, "hello\n", buffer.String())
}

func TestStreamFlushes(t *testing.T) {
	t.Parallel()

	destination := &flushTracker{}
	stream := NewStream(destination)

	require.NoError(t, stream.Write([]byte("x"), hqgologgerlevel.LevelInfo))
	assert.True(t, destination.flushed)
	assert.Equal(t, "x\n", destination.String())
}

func TestStreamNilUsesDiscard(t *testing.T) {
	t.Parallel()

	stream := NewStream(nil)

	require.NoError(t, stream.Write([]byte("x"), hqgologgerlevel.LevelInfo))
	require.NoError(t, stream.Close())
}

func TestStreamShortWriteAndError(t *testing.T) {
	t.Parallel()

	require.ErrorIs(t, NewStream(shortWriter{}).Write([]byte("x"), hqgologgerlevel.LevelInfo), io.ErrShortWrite)

	expected := errors.New("write failed")

	require.ErrorIs(t, NewStream(failingWriter{err: expected}).Write([]byte("x"), hqgologgerlevel.LevelInfo), expected)
}

func TestStreamClose(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	stream := NewStream(&buffer)

	require.NoError(t, stream.Close())
	require.NoError(t, stream.Close())
	require.ErrorIs(t, stream.Write([]byte("x"), hqgologgerlevel.LevelInfo), ErrClosed)
}

func TestStreamNilReceivers(t *testing.T) {
	t.Parallel()

	var stream *Stream

	require.ErrorIs(t, stream.Write([]byte("x"), hqgologgerlevel.LevelInfo), ErrClosed)
	require.NoError(t, stream.Close())

	zero := &Stream{}

	require.ErrorIs(t, zero.Write([]byte("x"), hqgologgerlevel.LevelInfo), ErrClosed)
	require.NoError(t, zero.Close())
}

func TestStreamCloseStdout(t *testing.T) {
	t.Parallel()

	stream := &Stream{writer: os.Stdout}

	require.NoError(t, stream.Close())
	require.ErrorIs(t, stream.Write([]byte("x"), hqgologgerlevel.LevelInfo), ErrClosed)
}

func TestStreamCloseStderr(t *testing.T) {
	t.Parallel()

	stream := &Stream{writer: os.Stderr}

	require.NoError(t, stream.Close())
}

type closingWriter struct {
	closed bool
	err    error
}

func (w *closingWriter) Write([]byte) (int, error) { return 0, nil }
func (w *closingWriter) Close() error {
	w.closed = true

	return w.err
}

func TestStreamCloseClosesUnderlying(t *testing.T) {
	t.Parallel()

	underlying := &closingWriter{}
	stream := NewStream(underlying)

	require.NoError(t, stream.Close())
	assert.True(t, underlying.closed)

	expected := errors.New("close failed")

	require.ErrorIs(t, NewStream(&closingWriter{err: expected}).Close(), expected)
}
