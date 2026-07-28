package logger

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"testing"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureWriter struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	levels   []hqgologgerlevels.Level
	closed   bool
	closeErr error
}

func (w *captureWriter) Write(data []byte, level hqgologgerlevels.Level) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf.Write(data)
	w.buf.WriteByte('\n')
	w.levels = append(w.levels, level)

	return nil
}

func (w *captureWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.closed = true

	return w.closeErr
}

func (w *captureWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.buf.String()
}

type failingWriter struct{}

func (failingWriter) Write([]byte, hqgologgerlevels.Level) error { return errors.New("sink down") }

func (failingWriter) Close() error { return nil }

func newTestLogger(w *captureWriter) *Logger {
	l := NewLogger()

	_ = l.SetLevel(hqgologgerlevels.LevelDebug) // LevelDebug is valid; cannot fail.

	l.SetFormatter(hqgologgerformatter.NewConsoleFormatter(&hqgologgerformatter.ConsoleFormatterConfiguration{
		IncludeLabel: true,
		Colorizer:    hqgologgerformatter.NewNoOpColorizer(),
	}))
	l.SetWriter(w)

	return l
}

func TestNewLoggerHasNoFormatterOrWriter(t *testing.T) {
	t.Parallel()

	l := NewLogger()
	require.NotNil(t, l)

	assert.NotPanics(t, func() {
		l.Info("dropped")
	})
}

func TestDefaultLabelsPerLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		log  func(l *Logger)
		want string
	}{
		{"error", func(l *Logger) { l.Error("m") }, "[ERR] m"},
		{"info", func(l *Logger) { l.Info("m") }, "[INF] m"},
		{"warn", func(l *Logger) { l.Warn("m") }, "[WRN] m"},
		{"debug", func(l *Logger) { l.Debug("m") }, "[DBG] m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &captureWriter{}
			l := newTestLogger(w)

			tt.log(l)
			assert.Contains(t, w.String(), tt.want)
		})
	}
}

func TestPrintHasNoDefaultLabel(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Print("plain")

	out := w.String()
	assert.Contains(t, out, "plain")
	assert.NotContains(t, out, "[")
	require.Len(t, w.levels, 1)
	assert.Equal(t, hqgologgerlevels.LevelSilent, w.levels[0])
}

func TestLevelFiltering(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	require.NoError(t, l.SetLevel(hqgologgerlevels.LevelError))

	l.Debug("dropped-debug")
	l.Warn("dropped-warn")
	l.Info("dropped-info")
	l.Error("kept-error")

	out := w.String()
	assert.NotContains(t, out, "dropped")
	assert.Contains(t, out, "kept-error")
}

func TestSetLevelRejectsInvalidLevel(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	require.NoError(t, l.SetLevel(hqgologgerlevels.LevelError))

	err := l.SetLevel(hqgologgerlevels.Level(-1))
	require.ErrorIs(t, err, hqgologgerlevels.ErrUnknownLevel)

	err = l.SetLevel(hqgologgerlevels.Level(99))
	require.ErrorIs(t, err, hqgologgerlevels.ErrUnknownLevel)

	// The rejected updates must leave the previous threshold in place.
	l.Info("dropped-info")
	l.Error("kept-error")

	out := w.String()
	assert.NotContains(t, out, "dropped")
	assert.Contains(t, out, "kept-error")
}

func TestLevelGetterReturnsCurrentThreshold(t *testing.T) {
	t.Parallel()

	l := newTestLogger(&captureWriter{})
	assert.Equal(t, hqgologgerlevels.LevelDebug, l.Level())

	require.NoError(t, l.SetLevel(hqgologgerlevels.LevelError))
	assert.Equal(t, hqgologgerlevels.LevelError, l.Level())
}

func TestEnabledReflectsThreshold(t *testing.T) {
	t.Parallel()

	l := newTestLogger(&captureWriter{})

	require.NoError(t, l.SetLevel(hqgologgerlevels.LevelError))

	assert.True(t, l.Enabled(hqgologgerlevels.LevelFatal))
	assert.True(t, l.Enabled(hqgologgerlevels.LevelSilent))
	assert.True(t, l.Enabled(hqgologgerlevels.LevelError))
	assert.False(t, l.Enabled(hqgologgerlevels.LevelInfo))
	assert.False(t, l.Enabled(hqgologgerlevels.LevelWarn))
	assert.False(t, l.Enabled(hqgologgerlevels.LevelDebug))
}

func TestLoggerCloseClosesWriter(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	require.NoError(t, l.Close())
	assert.True(t, w.closed)
}

func TestLoggerClosePropagatesWriterError(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	w := &captureWriter{closeErr: closeErr}
	l := newTestLogger(w)

	require.ErrorIs(t, l.Close(), closeErr)
	assert.True(t, w.closed)
}

func TestLoggerCloseWithoutWriterIsNoOp(t *testing.T) {
	t.Parallel()

	l := NewLogger()

	require.NoError(t, l.Close())
}

func TestLevelSilentThresholdSuppressesAll(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	require.NoError(t, l.SetLevel(hqgologgerlevels.LevelSilent))

	l.Error("err")
	l.Info("info")
	l.Print("printed")

	out := w.String()
	assert.NotContains(t, out, "err")
	assert.NotContains(t, out, "info")
	assert.Contains(t, out, "printed")
}

func TestWithLabelOverridesDefault(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Info("m", WithLabel("CUSTOM"))
	assert.Contains(t, w.String(), "[CUSTOM] m")
}

func TestWithoutLabelSuppressesLabel(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Info("m", WithoutLabel())

	out := w.String()
	assert.Contains(t, out, "m")
	assert.NotContains(t, out, "[INF]")
	assert.NotContains(t, out, "[]")
}

func TestWithStringMetadata(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Info("m", WithString("request_id", "12345"))
	assert.Contains(t, w.String(), "request_id=12345")
}

func TestWithValueMetadata(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Info("m", WithValue("count", 7))
	assert.Contains(t, w.String(), "count=7")
}

func TestWithErrorMetadata(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Error("failed", WithError(errors.New("boom")))

	out := w.String()
	assert.Contains(t, out, "[ERR] failed")
	assert.Contains(t, out, "boom")
	assert.NotContains(t, out, "error=")
}

func TestWithoutTimestamp(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := NewLogger()

	_ = l.SetLevel(hqgologgerlevels.LevelDebug) // LevelDebug is valid; cannot fail.

	l.SetFormatter(hqgologgerformatter.NewConsoleFormatter(&hqgologgerformatter.ConsoleFormatterConfiguration{
		IncludeTimestamp: true,
		TimestampFormat:  "2006",
		IncludeLabel:     true,
		Colorizer:        hqgologgerformatter.NewNoOpColorizer(),
	}))
	l.SetWriter(w)

	l.Info("m", WithoutTimestamp())

	assert.Equal(t, "[INF] m\n", w.String())
}

func TestMultipleOptionsCombine(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Info("started",
		WithLabel("START"),
		WithString("app", "my-app"),
		WithValue("pid", 99),
	)

	out := w.String()
	assert.Contains(t, out, "[START] started")
	assert.Contains(t, out, "app=my-app")
	assert.Contains(t, out, "pid=99")
}

func TestCustomOptionFunc(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Info("m", func(e *Event) { e.SetString("custom", "yes") })

	assert.Contains(t, w.String(), "custom=yes")
}

func TestLogDirectEvent(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)

	l.Log(NewEvent(
		WithLevel(hqgologgerlevels.LevelWarn),
		WithMessage("direct"),
		WithString("origin", "test"),
	))

	assert.Contains(t, w.String(), "[WRN] direct origin=test")
}

func TestLogNilEventIsNoOp(t *testing.T) {
	t.Parallel()

	l := newTestLogger(&captureWriter{})

	assert.NotPanics(t, func() {
		l.Log(nil)
	})
}

func TestWriteErrorIsSwallowed(t *testing.T) {
	t.Parallel()

	l := NewLogger()

	_ = l.SetLevel(hqgologgerlevels.LevelDebug) // LevelDebug is valid; cannot fail.

	l.SetFormatter(hqgologgerformatter.NewConsoleFormatter(nil))
	l.SetWriter(failingWriter{})

	assert.NotPanics(t, func() {
		l.Info("m")
	})
}

func TestSetWriterReplacesSink(t *testing.T) {
	t.Parallel()

	first := &captureWriter{}
	second := &captureWriter{}
	l := newTestLogger(first)

	l.Info("one")
	l.SetWriter(second)
	l.Info("two")

	assert.Contains(t, first.String(), "one")
	assert.NotContains(t, first.String(), "two")
	assert.Contains(t, second.String(), "two")
}

func TestConcurrentLogAndReconfigure(t *testing.T) {
	t.Parallel()

	l := newTestLogger(&captureWriter{})

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			for range 200 {
				l.Info("concurrent", WithString("k", "v"))
			}
		})
	}

	wg.Go(func() {
		for range 200 {
			l.SetWriter(&captureWriter{})
			l.SetFormatter(hqgologgerformatter.NewConsoleFormatter(nil))
			_ = l.SetLevel(hqgologgerlevels.LevelDebug) // LevelDebug is valid; cannot fail.
		}
	})

	wg.Wait()
}

func TestFatalExits(t *testing.T) {
	if os.Getenv("HQLOGGER_CRASH") == "1" {
		Fatal("fatal boom")

		return
	}

	t.Parallel()

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=TestFatalExits") //nolint:gosec

	cmd.Env = append(os.Environ(), "HQLOGGER_CRASH=1")

	err := cmd.Run()

	var exitErr *exec.ExitError

	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
}

func TestFatalExitsWhenUnconfigured(t *testing.T) {
	if os.Getenv("HQLOGGER_CRASH_UNCONFIGURED") == "1" {
		NewLogger().Fatal("fatal without formatter or writer")

		return
	}

	t.Parallel()

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=TestFatalExitsWhenUnconfigured") //nolint:gosec

	cmd.Env = append(os.Environ(), "HQLOGGER_CRASH_UNCONFIGURED=1")

	err := cmd.Run()

	var exitErr *exec.ExitError

	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
}

func BenchmarkInfo(b *testing.B) {
	l := NewLogger()

	_ = l.SetLevel(hqgologgerlevels.LevelDebug) // LevelDebug is valid; cannot fail.

	l.SetFormatter(hqgologgerformatter.NewConsoleFormatter(hqgologgerformatter.DefaultConsoleFormatterConfig()))
	l.SetWriter(hqgologgerwriter.NewConsoleWriter(&hqgologgerwriter.ConsoleWriterConfiguration{
		ForceStdout: true,
		Stdout:      io.Discard,
	}))

	b.ReportAllocs()

	for b.Loop() {
		l.Info("request completed", WithString("method", "GET"), WithValue("status", 200))
	}
}
