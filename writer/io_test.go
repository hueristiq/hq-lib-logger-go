package writer

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

type failingIOWriter struct {
	err error
}

func (f *failingIOWriter) Write([]byte) (int, error) { return 0, f.err }

type flushTracker struct {
	bytes.Buffer

	flushes int
}

func (f *flushTracker) Flush() error {
	f.flushes++

	return nil
}

func TestIOWriterWritesWithNewline(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	w := NewIOWriter(&buf)

	require.NoError(t, w.Write([]byte("hello"), hqgologgerlevels.LevelInfo))
	assert.Equal(t, "hello\n", buf.String())
}

func TestIOWriterWritesEveryLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	w := NewIOWriter(&buf)

	for _, level := range []hqgologgerlevels.Level{
		hqgologgerlevels.LevelFatal,
		hqgologgerlevels.LevelSilent,
		hqgologgerlevels.LevelError,
		hqgologgerlevels.LevelInfo,
		hqgologgerlevels.LevelWarn,
		hqgologgerlevels.LevelDebug,
	} {
		require.NoError(t, w.Write([]byte("x"), level))
	}

	assert.Equal(t, "x\nx\nx\nx\nx\nx\n", buf.String())
}

func TestIOWriterWriteDoesNotTouchSpareCapacity(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	w := NewIOWriter(&buf)

	data := make([]byte, 7, 8)
	copy(data, "payload")

	require.NoError(t, w.Write(data, hqgologgerlevels.LevelInfo))
	assert.Equal(t, "payload\n", buf.String())
	assert.Equal(t, byte(0), data[:cap(data)][cap(data)-1], "spare capacity must stay untouched")
}

func TestIOWriterPropagatesWriteError(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("disk full")
	w := NewIOWriter(&failingIOWriter{err: writeErr})

	require.ErrorIs(t, w.Write([]byte("x"), hqgologgerlevels.LevelInfo), writeErr)
}

func TestIOWriterFlushesWhenSupported(t *testing.T) {
	t.Parallel()

	f := &flushTracker{}
	w := NewIOWriter(f)

	require.NoError(t, w.Write([]byte("x"), hqgologgerlevels.LevelInfo))
	assert.Equal(t, 1, f.flushes)
	assert.Equal(t, "x\n", f.String())
}

func TestIOWriterCloseClosesUnderlying(t *testing.T) {
	t.Parallel()

	c := &trackingCloser{}
	w := NewIOWriter(c)

	require.NoError(t, w.Close())
	assert.True(t, c.closed)
}

func TestIOWriterCloseNonCloserIsNoOp(t *testing.T) {
	t.Parallel()

	w := NewIOWriter(&bytes.Buffer{})

	require.NoError(t, w.Close())
}

func TestIOWriterNilFallsBackToDiscard(t *testing.T) {
	t.Parallel()

	w := NewIOWriter(nil)
	require.NotNil(t, w)

	require.NoError(t, w.Write([]byte("x"), hqgologgerlevels.LevelInfo))
	require.NoError(t, w.Close())
}

func TestIOWriterConcurrentWrites(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	w := NewIOWriter(&buf)

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			for range 100 {
				_ = w.Write([]byte("x"), hqgologgerlevels.LevelInfo)
			}
		})
	}

	wg.Wait()

	assert.Equal(t, 800, bytes.Count(buf.Bytes(), []byte("x\n")))
}

func TestIOWriterImplementsWriter(t *testing.T) {
	t.Parallel()

	var _ Writer = NewIOWriter(nil)
}

type errorFlusher struct {
	bytes.Buffer

	err error
}

func (f *errorFlusher) Flush() error { return f.err }

func TestIOWriterFlushErrorPropagates(t *testing.T) {
	t.Parallel()

	flushErr := errors.New("flush failed")
	w := NewIOWriter(&errorFlusher{err: flushErr})

	require.ErrorIs(t, w.Write([]byte("x"), hqgologgerlevels.LevelInfo), flushErr)
}

func TestIOWriterClosePropagatesError(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	w := NewIOWriter(&failingCloser{err: closeErr})

	require.ErrorIs(t, w.Close(), closeErr)
}
