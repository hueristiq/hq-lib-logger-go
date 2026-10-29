package writer

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-logger-go/levels"
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

	require.NoError(t, stream.Write([]byte("hello"), levels.LevelInfo))
	assert.Equal(t, "hello\n", buffer.String())
}

func TestStreamFlushes(t *testing.T) {
	t.Parallel()

	destination := &flushTracker{}
	stream := NewStream(destination)

	require.NoError(t, stream.Write([]byte("x"), levels.LevelInfo))
	assert.True(t, destination.flushed)
	assert.Equal(t, "x\n", destination.String())
}

func TestStreamNilUsesDiscard(t *testing.T) {
	t.Parallel()

	stream := NewStream(nil)

	require.NoError(t, stream.Write([]byte("x"), levels.LevelInfo))
	require.NoError(t, stream.Close())
}

func TestStreamShortWriteAndError(t *testing.T) {
	t.Parallel()

	require.ErrorIs(t, NewStream(shortWriter{}).Write([]byte("x"), levels.LevelInfo), io.ErrShortWrite)

	expected := errors.New("write failed")

	require.ErrorIs(t, NewStream(failingWriter{err: expected}).Write([]byte("x"), levels.LevelInfo), expected)
}

func TestStreamClose(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	stream := NewStream(&buffer)

	require.NoError(t, stream.Close())
	require.NoError(t, stream.Close())
	require.ErrorIs(t, stream.Write([]byte("x"), levels.LevelInfo), ErrClosed)
}
