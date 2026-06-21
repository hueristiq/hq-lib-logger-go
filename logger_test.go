package logger

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"sync"
	"testing"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureWriter struct {
	mu     sync.Mutex
	buf    bytes.Buffer
	levels []hqgologgerlevels.Level
}

func (w *captureWriter) Write(data []byte, level hqgologgerlevels.Level) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf.Write(data)
	w.buf.WriteByte('\n')
	w.levels = append(w.levels, level)

	return nil
}

func (w *captureWriter) Close() error { return nil }

func (w *captureWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.buf.String()
}

func newTestLogger(w *captureWriter) *Logger {
	l := NewLogger()

	l.SetLevel(hqgologgerlevels.LevelDebug)
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
	l.SetLevel(hqgologgerlevels.LevelError)

	l.Debug("dropped-debug")
	l.Warn("dropped-warn")
	l.Info("dropped-info")
	l.Error("kept-error")

	out := w.String()
	assert.NotContains(t, out, "dropped")
	assert.Contains(t, out, "kept-error")
}

func TestLevelSilentThresholdSuppressesAll(t *testing.T) {
	t.Parallel()

	w := &captureWriter{}
	l := newTestLogger(w)
	l.SetLevel(hqgologgerlevels.LevelSilent)

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
	l.SetLevel(hqgologgerlevels.LevelDebug)
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
			l.SetLevel(hqgologgerlevels.LevelDebug)
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
