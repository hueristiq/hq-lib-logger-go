package writer

import (
	"errors"
	"testing"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingWriter struct {
	writes   [][]byte
	levels   []hqgologgerlevels.Level
	writeErr error
	closeErr error
	closed   bool
}

func (r *recordingWriter) Write(data []byte, level hqgologgerlevels.Level) error {
	r.writes = append(r.writes, append([]byte(nil), data...))
	r.levels = append(r.levels, level)

	return r.writeErr
}

func (r *recordingWriter) Close() error {
	r.closed = true

	return r.closeErr
}

func TestMultiWriterFansOut(t *testing.T) {
	t.Parallel()

	a := &recordingWriter{}
	b := &recordingWriter{}

	m := NewMultiWriter(a, b)

	require.NoError(t, m.Write([]byte("payload"), hqgologgerlevels.LevelInfo))

	require.Len(t, a.writes, 1)
	require.Len(t, b.writes, 1)
	assert.Equal(t, "payload", string(a.writes[0]))
	assert.Equal(t, "payload", string(b.writes[0]))
	assert.Equal(t, hqgologgerlevels.LevelInfo, a.levels[0])
	assert.Equal(t, hqgologgerlevels.LevelInfo, b.levels[0])
}

func TestMultiWriterFiltersNilWriters(t *testing.T) {
	t.Parallel()

	a := &recordingWriter{}

	m := NewMultiWriter(nil, a, nil)

	require.NoError(t, m.Write([]byte("x"), hqgologgerlevels.LevelDebug))
	assert.Len(t, a.writes, 1)
}

func TestMultiWriterEmptyIsNoOp(t *testing.T) {
	t.Parallel()

	m := NewMultiWriter()

	require.NoError(t, m.Write([]byte("x"), hqgologgerlevels.LevelInfo))
	require.NoError(t, m.Close())
}

func TestMultiWriterWriteAttemptsAllDespiteError(t *testing.T) {
	t.Parallel()

	failErr := errors.New("disk full")
	a := &recordingWriter{writeErr: failErr}
	b := &recordingWriter{}

	m := NewMultiWriter(a, b)

	err := m.Write([]byte("x"), hqgologgerlevels.LevelInfo)

	assert.Len(t, a.writes, 1)
	assert.Len(t, b.writes, 1)

	// The later success must not mask the earlier failure.
	require.ErrorIs(t, err, failErr)
}

func TestMultiWriterWriteReturnsLastError(t *testing.T) {
	t.Parallel()

	lastErr := errors.New("second failed")
	a := &recordingWriter{writeErr: errors.New("first failed")}
	b := &recordingWriter{writeErr: lastErr}

	m := NewMultiWriter(a, b)

	err := m.Write([]byte("x"), hqgologgerlevels.LevelInfo)
	require.ErrorIs(t, err, lastErr)
}

func TestMultiWriterCloseClosesAll(t *testing.T) {
	t.Parallel()

	a := &recordingWriter{}
	b := &recordingWriter{}

	m := NewMultiWriter(a, b)

	require.NoError(t, m.Close())
	assert.True(t, a.closed)
	assert.True(t, b.closed)
}

func TestMultiWriterCloseReturnsLastError(t *testing.T) {
	t.Parallel()

	lastErr := errors.New("close b failed")
	a := &recordingWriter{closeErr: errors.New("close a failed")}
	b := &recordingWriter{closeErr: lastErr}

	m := NewMultiWriter(a, b)

	err := m.Close()
	require.ErrorIs(t, err, lastErr)
	assert.True(t, a.closed)
	assert.True(t, b.closed)
}

func TestMultiWriterCloseErrorNotMaskedByLaterSuccess(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close a failed")
	a := &recordingWriter{closeErr: closeErr}
	b := &recordingWriter{}

	m := NewMultiWriter(a, b)

	err := m.Close()
	require.ErrorIs(t, err, closeErr)
	assert.True(t, a.closed)
	assert.True(t, b.closed)
}

func TestMultiWriterImplementsWriter(t *testing.T) {
	t.Parallel()

	var _ Writer = NewMultiWriter()
}
