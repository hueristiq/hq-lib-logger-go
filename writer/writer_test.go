package writer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

type recordingWriter struct {
	data   []byte
	levels []hqgologgerlevels.Level
	closed bool
	err    error
}

func (r *recordingWriter) Write(data []byte, level hqgologgerlevels.Level) error {
	if r.err != nil {
		return r.err
	}

	r.data = append(r.data, data...)
	r.levels = append(r.levels, level)

	return nil
}

func (r *recordingWriter) Close() error {
	r.closed = true

	return nil
}

type syncWriter struct {
	recordingWriter
	synced bool
}

func (s *syncWriter) Sync() error {
	s.synced = true

	return nil
}

func TestMultiWriterForwards(t *testing.T) {
	t.Parallel()

	first := &recordingWriter{}
	second := &recordingWriter{}
	multi := NewMulti(first, nil, second)

	require.NoError(t, multi.Write([]byte("x"), hqgologgerlevels.LevelInfo))
	assert.Equal(t, []byte("x"), first.data)
	assert.Equal(t, []byte("x"), second.data)
	assert.Equal(t, []hqgologgerlevels.Level{hqgologgerlevels.LevelInfo}, first.levels)
}

func TestMultiWriterJoinsErrors(t *testing.T) {
	t.Parallel()

	firstErr := errors.New("first")
	secondErr := errors.New("second")
	multi := NewMulti(
		&recordingWriter{err: firstErr},
		&recordingWriter{err: secondErr},
	)

	err := multi.Write([]byte("x"), hqgologgerlevels.LevelInfo)

	require.ErrorIs(t, err, firstErr)
	require.ErrorIs(t, err, secondErr)
}

func TestMultiWriterSyncAndClose(t *testing.T) {
	t.Parallel()

	first := &syncWriter{}
	second := &recordingWriter{}
	multi := NewMulti(first, second)

	require.NoError(t, multi.Sync())
	assert.True(t, first.synced)

	require.NoError(t, multi.Close())
	assert.True(t, first.closed)
	assert.True(t, second.closed)

	require.NoError(t, multi.Close())
	require.ErrorIs(t, multi.Write([]byte("x"), hqgologgerlevels.LevelInfo), ErrClosed)
	require.ErrorIs(t, multi.Sync(), ErrClosed)
}

func TestMultiWriterNilReceiver(t *testing.T) {
	t.Parallel()

	var multi *MultiWriter

	require.ErrorIs(t, multi.Write([]byte("x"), hqgologgerlevels.LevelInfo), ErrClosed)
	require.ErrorIs(t, multi.Sync(), ErrClosed)
	require.NoError(t, multi.Close())
}

func TestMultiWriterSyncPlainWriter(t *testing.T) {
	t.Parallel()

	multi := NewMulti(&recordingWriter{})

	require.NoError(t, multi.Sync())
}

type failingSyncWriter struct {
	recordingWriter
	syncErr error
}

func (f *failingSyncWriter) Sync() error { return f.syncErr }

type flusherWriter struct {
	recordingWriter
	flushed  bool
	flushErr error
}

func (f *flusherWriter) Flush() error {
	f.flushed = true

	return f.flushErr
}

type failingCloseWriter struct {
	recordingWriter
	closeErr error
}

func (f *failingCloseWriter) Close() error {
	f.closed = true

	return f.closeErr
}

func TestMultiWriterSyncErrors(t *testing.T) {
	t.Parallel()

	syncErr := errors.New("sync failed")
	flushErr := errors.New("flush failed")

	multi := NewMulti(
		&failingSyncWriter{syncErr: syncErr},
		&flusherWriter{flushErr: flushErr},
	)

	err := multi.Sync()

	require.ErrorIs(t, err, syncErr)
	require.ErrorIs(t, err, flushErr)

	multi = NewMulti(&flusherWriter{})

	require.NoError(t, multi.Sync())
}

func TestMultiWriterCloseError(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	multi := NewMulti(&failingCloseWriter{closeErr: closeErr})

	require.ErrorIs(t, multi.Close(), closeErr)
}
