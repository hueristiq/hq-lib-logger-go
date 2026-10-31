//nolint:testpackage // tests exercise unexported behavior.
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
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

// Reset clears the captured output for the next assertions.
func (w *captureWriter) Reset() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.buffer.Reset()
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

	require.ErrorIs(t, l.SetLevel(Level(99)), hqgologgerlevel.ErrUnknownLevel)
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
	l := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)))

	l.LogRecord(hqgologgerformatter.Log{Level: LevelInfo, Message: "record"})

	var decoded map[string]any

	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(sink.String())), &decoded))
	assert.Equal(t, "record", decoded["message"])
	assert.NotContains(t, decoded, "time")
}

func TestLogRecordCallerAndName(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)))

	l.LogRecord(hqgologgerformatter.Log{
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
		WithHooks(HookFunc(func(hqgologgerformatter.Log) error {
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

func benchmarkEmit(b *testing.B, f hqgologgerformatter.Formatter) {
	b.Helper()

	l := New(WithFormatter(f), WithWriter(hqgologgerwriter.NewStream(io.Discard)), WithLevel(LevelTrace))

	b.ReportAllocs()
	b.ResetTimer()

	for i := range b.N {
		l.Info("message", String("key", "value"), Int("n", i))
	}
}

func BenchmarkEmitConsole(b *testing.B) {
	benchmarkEmit(b, hqgologgerformatter.NewConsole(&hqgologgerformatter.ConsoleConfig{Timestamp: false, Label: true}))
}

func BenchmarkEmitJSON(b *testing.B) {
	benchmarkEmit(b, hqgologgerformatter.NewJSON(nil))
}

func ExampleNew() {
	var buffer bytes.Buffer

	l := New(
		WithWriter(hqgologgerwriter.NewStream(&buffer)),
		WithFormatter(hqgologgerformatter.NewConsole(&hqgologgerformatter.ConsoleConfig{Timestamp: false, Label: true})),
	)

	l.Info("hello", String("service", "api"))

	fmt.Print(buffer.String())
	// Output:
	// [INF] hello service=api
}

func ExampleLogger_Infof() {
	var buffer bytes.Buffer

	l := New(
		WithWriter(hqgologgerwriter.NewStream(&buffer)),
		WithFormatter(hqgologgerformatter.NewConsole(&hqgologgerformatter.ConsoleConfig{Timestamp: false, Label: true})),
	)

	l.Infof("hello %s", "world")

	fmt.Print(buffer.String())
	// Output:
	// [INF] hello world
}

func TestFieldConstructors(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	err := errors.New("boom")

	fields := []Field{
		String("string", "v"),
		Int("int", 1),
		Int64("int64", 2),
		Uint("uint", 3),
		Uint64("uint64", 4),
		Float64("float64", 1.5),
		Bool("bool", true),
		Duration("duration", time.Second),
		Time("time", now),
		Err("err", err),
		Any("any", "anything"),
	}

	for _, field := range fields {
		assert.NotEmpty(t, field.Key())
		assert.NotNil(t, field.Value())
	}
}

func TestLogLevels(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithLevel(LevelTrace))

	l.Log(Level(99), "dropped")
	assert.Empty(t, sink.String())

	l.Log(LevelDebug, "kept")
	assert.Contains(t, sink.String(), "kept")
}

func TestFatalExits(t *testing.T) {
	t.Parallel()

	if os.Getenv("HQ_LOGGER_FATAL_HELPER") == "1" {
		New().Fatal("boom")

		return
	}

	if os.Getenv("HQ_LOGGER_FATAL_HELPER") == "2" {
		New().Fatalf("boom %d", 1)

		return
	}

	for _, helperEnv := range []string{"1", "2"} {
		//nolint:gosec // The test re-executes the test binary itself with a static argument list.
		command := exec.CommandContext(context.Background(), os.Args[0], "-test.run=TestFatalExits")

		command.Env = append(os.Environ(), "HQ_LOGGER_FATAL_HELPER="+helperEnv)

		err := command.Run()

		var exitErr *exec.ExitError

		require.ErrorAs(t, err, &exitErr)
		assert.Equal(t, 1, exitErr.ExitCode())
	}
}

func TestPrintfSugar(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)), WithLevel(LevelTrace))

	l.Infof("hello %s %d", "world", 42)
	l.Warnf("warn %d", 1)
	l.Errorf("error %d", 2)
	l.Debugf("debug %d", 3)
	l.Tracef("trace %d", 4)
	l.Logf(LevelDebug, "logf %d", 5)

	output := sink.String()
	assert.Contains(t, output, "hello world 42")
	assert.Contains(t, output, "warn 1")
	assert.Contains(t, output, "error 2")
	assert.Contains(t, output, "debug 3")
	assert.Contains(t, output, "trace 4")
	assert.Contains(t, output, "logf 5")
}

func TestLogRecordVariants(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)))

	l.LogRecord(hqgologgerformatter.Log{Level: Level(99), Message: "dropped"})
	assert.Empty(t, sink.String())

	prefixed := l.WithPrefix("[p] ")
	prefixed.LogRecord(hqgologgerformatter.Log{Level: LevelInfo, Message: "prefixed"})

	decoded := decodeRecord(t, sink)
	assert.Equal(t, "[p] prefixed", decoded["message"])

	sink.Reset()

	named := l.Named("api")
	named.LogRecord(hqgologgerformatter.Log{Level: LevelInfo, Message: "named"})

	decoded = decodeRecord(t, sink)
	assert.Equal(t, "api", decoded["logger"])

	sink.Reset()

	withCaller := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)), WithCallerCapture(true))
	withCaller.LogRecord(hqgologgerformatter.Log{Level: LevelInfo, Message: "caller"})

	decoded = decodeRecord(t, sink)
	assert.Contains(t, decoded["caller"], "logger_test.go:")

	sink.Reset()

	withFields := l.With(String("base", "1"))
	withFields.LogRecord(hqgologgerformatter.Log{
		Level:   LevelInfo,
		Message: "fields",
		Fields:  []Field{Int("extra", 2)},
	})

	decoded = decodeRecord(t, sink)
	assert.Equal(t, "1", decoded["base"])
	assert.InDelta(t, 2.0, decoded["extra"], 0)
}

func TestWithSamplerNilDisables(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithSampler(SamplerFunc(func(Level) bool { return false })))
	l.Info("dropped")
	assert.Empty(t, sink.String())

	sink.Reset()

	unsampled := New(WithWriter(sink), WithSampler(SamplerFunc(nil)))
	unsampled.Info("kept")
	assert.Contains(t, sink.String(), "kept")
}

func TestGetCoreConcurrentFirstUse(t *testing.T) {
	t.Parallel()

	var l Logger

	var wg sync.WaitGroup

	for range 8 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_ = l.getCore()
		}()
	}

	wg.Wait()

	assert.Equal(t, LevelInfo, l.Level())
}

func TestChildViewVariants(t *testing.T) {
	t.Parallel()

	l := New()

	assert.Empty(t, l.Named("").name)

	child := l.Named("api").Named("v1")
	assert.Equal(t, "api.v1", child.name)

	assert.Nil(t, l.With().fields)

	withFields := l.With(String("k", "v"))
	assert.NotNil(t, withFields.fields)

	assert.Empty(t, l.WithPrefix("").prefix)
	assert.Equal(t, "[a] [b] ", l.WithPrefix("[a] ").WithPrefix("[b] ").prefix)
}

type noSyncer struct{}

func (noSyncer) Write([]byte, Level) error { return nil }
func (noSyncer) Close() error              { return nil }

func TestSyncWithoutSyncer(t *testing.T) {
	t.Parallel()

	l := New(WithWriter(noSyncer{}))

	require.NoError(t, l.Sync())
}

func TestSyncAndCloseNilWriterCore(t *testing.T) {
	t.Parallel()

	core := defaultCore()
	core.writer = nil

	l := &Logger{}
	l.core.Store(core)

	require.NoError(t, l.Sync())
	require.NoError(t, l.Close())
}

type formatOnlyFormatter struct{}

func (formatOnlyFormatter) Format(hqgologgerformatter.Log) ([]byte, error) {
	return []byte("formatted"), nil
}

func TestEmitNonAppenderFormatter(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(formatOnlyFormatter{}))
	l.Info("hello")

	assert.Contains(t, sink.String(), "formatted")
}

func TestEmitNilFormatterDrops(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}

	core := defaultCore()
	core.writer = sink
	core.formatter = nil

	l := &Logger{}
	l.core.Store(core)

	l.LogRecord(hqgologgerformatter.Log{Level: LevelInfo, Message: "dropped"})

	assert.Empty(t, sink.String())
}

func TestMergeFields(t *testing.T) {
	t.Parallel()

	assert.Nil(t, mergeFields(nil, nil))
	assert.Len(t, mergeFields(nil, []Field{String("a", "1")}), 1)
	assert.Len(t, mergeFields([]Field{String("a", "1")}, nil), 1)
	assert.Len(t, mergeFields([]Field{String("a", "1")}, []Field{String("b", "2")}), 2)
}

func TestHooksFire(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}

	var hooked bool

	l := New(
		WithWriter(sink),
		WithHooks(HookFunc(func(hqgologgerformatter.Log) error {
			hooked = true

			return nil
		})),
	)

	l.Info("hello")

	assert.True(t, hooked)
	assert.Contains(t, sink.String(), "hello")
}

type flushWriter struct{}

func (flushWriter) Write([]byte, Level) error { return nil }
func (flushWriter) Close() error              { return nil }
func (flushWriter) Flush() error              { return nil }

func TestSyncFlushesWriter(t *testing.T) {
	t.Parallel()

	l := New(WithWriter(flushWriter{}))

	require.NoError(t, l.Sync())
}
