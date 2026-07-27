package writer

import (
	"bytes"
	"errors"
	"testing"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsoleWriterNilUsesDefaults(t *testing.T) {
	t.Parallel()

	w := NewConsoleWriter(nil)
	require.NotNil(t, w)

	require.NoError(t, w.Write([]byte("x"), hqgologgerlevels.LevelSilent))
}

func TestDefaultConsoleWriterConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConsoleWriterConfig()
	require.NotNil(t, cfg)
	assert.False(t, cfg.ForceStderr)
	assert.False(t, cfg.ForceStdout)
	assert.False(t, cfg.DisableNewline)
	assert.Nil(t, cfg.Stdout)
	assert.Nil(t, cfg.Stderr)
}

func TestConsoleWriterRoutesSilentToStdout(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.NoError(t, w.Write([]byte("hello"), hqgologgerlevels.LevelSilent))

	assert.Equal(t, "hello\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestConsoleWriterRoutesNonSilentToStderr(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	for _, level := range []hqgologgerlevels.Level{
		hqgologgerlevels.LevelFatal,
		hqgologgerlevels.LevelError,
		hqgologgerlevels.LevelInfo,
		hqgologgerlevels.LevelWarn,
		hqgologgerlevels.LevelDebug,
	} {
		require.NoError(t, w.Write([]byte("e"), level))
	}

	assert.Empty(t, stdout.String())
	assert.Equal(t, "e\ne\ne\ne\ne\n", stderr.String())
}

func TestConsoleWriterForceStdout(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		ForceStdout: true,
		Stdout:      &stdout,
		Stderr:      &stderr,
	})

	require.NoError(t, w.Write([]byte("err"), hqgologgerlevels.LevelError))

	assert.Equal(t, "err\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestConsoleWriterForceStderr(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		ForceStderr: true,
		Stdout:      &stdout,
		Stderr:      &stderr,
	})

	require.NoError(t, w.Write([]byte("silent"), hqgologgerlevels.LevelSilent))

	assert.Equal(t, "silent\n", stderr.String())
	assert.Empty(t, stdout.String())
}

func TestConsoleWriterForceStderrWinsOverStdout(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		ForceStderr: true,
		ForceStdout: true,
		Stdout:      &stdout,
		Stderr:      &stderr,
	})

	require.NoError(t, w.Write([]byte("x"), hqgologgerlevels.LevelInfo))

	assert.Equal(t, "x\n", stderr.String())
	assert.Empty(t, stdout.String())
}

func TestConsoleWriterDisableNewline(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		ForceStdout:    true,
		DisableNewline: true,
		Stdout:         &stdout,
	})

	require.NoError(t, w.Write([]byte("a"), hqgologgerlevels.LevelInfo))
	require.NoError(t, w.Write([]byte("b"), hqgologgerlevels.LevelInfo))

	assert.Equal(t, "ab", stdout.String())
}

func TestConsoleWriterWriteDoesNotMutateDataLength(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		ForceStdout: true,
		Stdout:      &stdout,
	})

	data := []byte("payload")

	require.NoError(t, w.Write(data, hqgologgerlevels.LevelInfo))
	assert.Len(t, data, len("payload"))
	assert.Equal(t, "payload\n", stdout.String())
}

func TestConsoleWriterCloseDoesNotCloseOSStreams(t *testing.T) {
	t.Parallel()

	w := NewConsoleWriter(nil)

	require.NoError(t, w.Close())
}

type trackingCloser struct {
	bytes.Buffer

	closed bool
}

func (t *trackingCloser) Close() error {
	t.closed = true

	return nil
}

type failingCloser struct {
	bytes.Buffer

	err error
}

func (f *failingCloser) Close() error { return f.err }

func TestConsoleWriterCloseClosesCustomStreams(t *testing.T) {
	t.Parallel()

	out := &trackingCloser{}
	errStream := &trackingCloser{}

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		Stdout: out,
		Stderr: errStream,
	})

	require.NoError(t, w.Close())
	assert.True(t, out.closed)
	assert.True(t, errStream.closed)
}

func TestConsoleWriterCloseReturnsStreamError(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	out := &trackingCloser{}

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		Stdout: out,
		Stderr: &failingCloser{err: closeErr},
	})

	err := w.Close()

	require.ErrorIs(t, err, closeErr)
	assert.True(t, out.closed, "stdout must still be closed when stderr fails")
}

func TestConsoleImplementsWriter(t *testing.T) {
	t.Parallel()

	var _ Writer = NewConsoleWriter(nil)
}
