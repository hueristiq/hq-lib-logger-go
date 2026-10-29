package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-logger-go/formatter"
	"github.com/hueristiq/hq-lib-logger-go/levels"
	"github.com/hueristiq/hq-lib-logger-go/writer"
)

type captureWriter struct {
	mutex      sync.Mutex
	buffer     bytes.Buffer
	levels     []Level
	closed     bool
	syncs      int
	panicWrite bool
}

func (w *captureWriter) Write(data []byte, level Level) error {
	if w.panicWrite {
		panic("writer panic")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.buffer.Write(data)
	w.levels = append(w.levels, level)

	return nil
}

func (w *captureWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.closed = true

	return nil
}

func (w *captureWriter) Sync() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.syncs++

	return nil
}

func (w *captureWriter) String() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.buffer.String()
}

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink))

	assert.Equal(t, LevelInfo, l.Level())
	assert.True(t, l.Enabled(LevelInfo))
	assert.False(t, l.Enabled(LevelDebug))

	l.Info("hello", String("service", "api"))

	assert.Contains(t, sink.String(), "[INF] hello")
	assert.Contains(t, sink.String(), "service=api")
}

func TestLevelMethods(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithLevel(LevelTrace), WithWriter(sink))

	l.Trace("trace")
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")

	assert.Equal(t, []Level{LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError}, sink.levels)
}

func TestSetLevel(t *testing.T) {
	t.Parallel()

	l := New()

	require.ErrorIs(t, l.SetLevel(Level(99)), levels.ErrUnknownLevel)
	assert.Equal(t, LevelInfo, l.Level())

	require.NoError(t, l.SetLevel(LevelDebug))
	assert.Equal(t, LevelDebug, l.Level())
	assert.True(t, l.Enabled(LevelInfo))
	assert.False(t, l.Enabled(LevelTrace))
}

func TestChildViews(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink))

	child := l.Named("api").WithPrefix("[request] ").With(String("request_id", "42"))
	child.Info("started")

	output := sink.String()
	assert.Contains(t, output, "[INF] [api] [request] started")
	assert.Contains(t, output, "request_id=42")
}

func TestLogRecordOmitsZeroTime(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(formatter.NewJSON(nil)))

	l.LogRecord(formatter.Log{Level: LevelInfo, Message: "record"})

	var decoded map[string]any

	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(sink.String())), &decoded))
	assert.Equal(t, "record", decoded["message"])
	assert.NotContains(t, decoded, "time")
}

func TestLogRecordCallerAndName(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(formatter.NewJSON(nil)))

	l.LogRecord(formatter.Log{
		Level:   LevelInfo,
		Message: "record",
		Caller:  "main.go:1",
		Name:    "api",
	})

	var decoded map[string]any

	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(sink.String())), &decoded))
	assert.Equal(t, "main.go:1", decoded["caller"])
	assert.Equal(t, "api", decoded["logger"])
}

func TestAsWriter(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink))

	n, err := l.AsWriter().Write([]byte("plain\n"))

	require.NoError(t, err)
	assert.Equal(t, len("plain\n"), n)
	assert.Contains(t, sink.String(), "[INF] plain")
}

func TestHooksAndSampling(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}

	var hooked int

	l := New(
		WithWriter(sink),
		WithHooks(HookFunc(func(*formatter.Log) error {
			hooked++

			return nil
		})),
		WithSampler(SamplerFunc(func(Level) bool { return false })),
	)

	l.Info("dropped")

	assert.Empty(t, sink.String())
	assert.Zero(t, hooked)

	l = New(WithWriter(sink))
	l.Info("kept")
	assert.Contains(t, sink.String(), "kept")
}

func TestSyncAndClose(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink))

	require.NoError(t, l.Sync())
	assert.Equal(t, 1, sink.syncs)

	require.NoError(t, l.Close())
	assert.True(t, sink.closed)

	require.NoError(t, l.Close())
	l.Info("after close")
	assert.Empty(t, sink.String())
}

func TestPanickingWriterReleasesLock(t *testing.T) {
	t.Parallel()

	l := New(WithWriter(&captureWriter{panicWrite: true}))

	func() {
		defer func() {
			require.NotNil(t, recover(), "writer must panic")
		}()

		l.Info("boom")
	}()

	done := make(chan struct{})

	go func() {
		_ = l.SetLevel(LevelDebug)

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("SetLevel blocked after a callback panic")
	}
}

func TestZeroValueLogger(t *testing.T) {
	t.Parallel()

	var l Logger

	assert.Equal(t, LevelInfo, l.Level())
	assert.True(t, l.Enabled(LevelInfo))

	child := l.With(String("k", "v"))
	assert.NotNil(t, child)
}

func benchmarkEmit(b *testing.B, f formatter.Formatter) {
	b.Helper()

	l := New(WithFormatter(f), WithWriter(writer.NewStream(io.Discard)), WithLevel(LevelTrace))

	b.ReportAllocs()
	b.ResetTimer()

	for i := range b.N {
		l.Info("message", String("key", "value"), Int("n", i))
	}
}

func BenchmarkEmitConsole(b *testing.B) {
	benchmarkEmit(b, formatter.NewConsole(&formatter.ConsoleConfig{Timestamp: false, Label: true}))
}

func BenchmarkEmitJSON(b *testing.B) {
	benchmarkEmit(b, formatter.NewJSON(nil))
}

func ExampleNew() {
	var buffer bytes.Buffer

	l := New(
		WithWriter(writer.NewStream(&buffer)),
		WithFormatter(formatter.NewConsole(&formatter.ConsoleConfig{Timestamp: false, Label: true})),
	)

	l.Info("hello", String("service", "api"))

	fmt.Print(buffer.String())
	// Output:
	// [INF] hello service=api
}
