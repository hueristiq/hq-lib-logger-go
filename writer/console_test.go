package writer

import (
	"bytes"
	"io"
	"testing"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setStreams(w *Console, stdout, stderr io.Writer) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.stdout = stdout
	w.stderr = stderr
}

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
}

func TestConsoleWriterRoutesSilentToStdout(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(nil)
	setStreams(w, &stdout, &stderr)

	require.NoError(t, w.Write([]byte("hello"), hqgologgerlevels.LevelSilent))

	assert.Equal(t, "hello\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestConsoleWriterRoutesNonSilentToStderr(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(nil)
	setStreams(w, &stdout, &stderr)

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

	w := NewConsoleWriter(&ConsoleWriterConfiguration{ForceStdout: true})
	setStreams(w, &stdout, &stderr)

	require.NoError(t, w.Write([]byte("err"), hqgologgerlevels.LevelError))

	assert.Equal(t, "err\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestConsoleWriterForceStderr(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{ForceStderr: true})
	setStreams(w, &stdout, &stderr)

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
	})
	setStreams(w, &stdout, &stderr)

	require.NoError(t, w.Write([]byte("x"), hqgologgerlevels.LevelInfo))

	assert.Equal(t, "x\n", stderr.String())
	assert.Empty(t, stdout.String())
}

func TestConsoleWriterDisableNewline(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	w := NewConsoleWriter(&ConsoleWriterConfiguration{
		ForceStdout:    true,
		DisableNewline: true,
	})
	setStreams(w, &stdout, &stderr)

	require.NoError(t, w.Write([]byte("a"), hqgologgerlevels.LevelInfo))
	require.NoError(t, w.Write([]byte("b"), hqgologgerlevels.LevelInfo))

	assert.Equal(t, "ab", stdout.String())
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

func TestConsoleWriterCloseClosesCustomStreams(t *testing.T) {
	t.Parallel()

	out := &trackingCloser{}
	errStream := &trackingCloser{}

	w := NewConsoleWriter(nil)
	setStreams(w, out, errStream)

	require.NoError(t, w.Close())
	assert.True(t, out.closed)
	assert.True(t, errStream.closed)
}

func TestConsoleImplementsWriter(t *testing.T) {
	t.Parallel()

	var _ Writer = NewConsoleWriter(nil)
}
