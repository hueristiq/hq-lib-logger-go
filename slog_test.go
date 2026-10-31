//nolint:testpackage // tests exercise unexported behavior.
package logger

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
)

func decodeRecord(t *testing.T, sink *captureWriter) (decoded map[string]any) {
	t.Helper()

	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(sink.String())), &decoded))

	return decoded
}

func TestSlogHandlerBasic(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)))

	slog.New(l.SlogHandler()).Info("hello", "request_id", "42")

	decoded := decodeRecord(t, sink)

	assert.Equal(t, "info", decoded["level"])
	assert.Equal(t, "hello", decoded["message"])
	assert.Equal(t, "42", decoded["request_id"])
}

func TestSlogHandlerAttrsAndGroups(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)))

	handler := l.SlogHandler().
		WithAttrs([]slog.Attr{slog.String("service", "api")}).
		WithGroup("http")

	slog.New(handler).Info("hello", "method", "GET")

	decoded := decodeRecord(t, sink)

	assert.Equal(t, "api", decoded["service"])
	assert.Equal(t, "GET", decoded["http.method"])
}

func TestSlogHandlerInlineGroup(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)))

	slog.New(l.SlogHandler().WithGroup("g")).Info("hello",
		slog.Group("", slog.String("a", "1")),
		slog.String("b", "2"),
	)

	decoded := decodeRecord(t, sink)

	assert.Equal(t, "1", decoded["g.a"])
	assert.Equal(t, "2", decoded["g.b"])
	assert.NotContains(t, decoded, "g..a")
}

func TestSlogHandlerCaller(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(
		WithWriter(sink),
		WithFormatter(hqgologgerformatter.NewJSON(nil)),
		WithCallerCapture(true),
	)

	slog.New(l.SlogHandler()).Info("hello")

	decoded := decodeRecord(t, sink)

	assert.Contains(t, decoded["caller"], "slog_test.go:")
}

func TestSlogHandlerNilLogger(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, NewSlogHandler(nil))
}

func TestSlogHandlerWithGroupEmptyName(t *testing.T) {
	t.Parallel()

	handler := New().SlogHandler()

	assert.Same(t, handler, handler.WithGroup(""))
}

func TestSlogHandlerRecordPC(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(
		WithWriter(sink),
		WithFormatter(hqgologgerformatter.NewJSON(nil)),
		WithCallerCapture(true),
	)

	pc, _, _, _ := runtime.Caller(0)
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "hello", pc)

	require.NoError(t, l.SlogHandler().Handle(context.Background(), record))

	decoded := decodeRecord(t, sink)

	caller, ok := decoded["caller"].(string)
	require.True(t, ok)
	assert.Contains(t, caller, "slog_test.go:")
}

func TestSlogHandlerEnabled(t *testing.T) {
	t.Parallel()

	handler := New(WithLevel(LevelInfo)).SlogHandler()

	assert.True(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.False(t, handler.Enabled(context.Background(), slog.LevelDebug))
}

func TestSlogToLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   slog.Level
		want Level
	}{
		{slog.LevelDebug - 4, LevelTrace},
		{slog.LevelDebug, LevelDebug},
		{slog.LevelInfo, LevelInfo},
		{slog.LevelWarn, LevelWarn},
		{slog.LevelError, LevelError},
		{slog.LevelError + 4, LevelError},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, slogToLevel(tt.in))
	}
}

func TestAppendSlogAttrKinds(t *testing.T) {
	t.Parallel()

	attrs := []slog.Attr{
		slog.String("string", "s"),
		slog.Int64("int64", 1),
		slog.Uint64("uint64", 2),
		slog.Float64("float64", 1.5),
		slog.Bool("bool", true),
		slog.Duration("duration", time.Second),
		slog.Time("time", time.Unix(0, 0)),
		slog.Any("any", "v"),
		slog.Any("valuer", slog.StringValue("lv")),
	}

	var fields []Field

	for _, attr := range attrs {
		fields = appendSlogAttr(fields, nil, attr)
	}

	assert.Len(t, fields, len(attrs))

	grouped := appendSlogAttr(nil, []string{"g"}, slog.String("k", "v"))
	require.Len(t, grouped, 1)
	assert.Equal(t, "g.k", grouped[0].Key())

	emptyKey := appendSlogAttr([]Field{String("keep", "1")}, nil, slog.Attr{Key: "", Value: slog.StringValue("x")})
	require.Len(t, emptyKey, 1)
	assert.Equal(t, "keep", emptyKey[0].Key())

	namedGroup := appendSlogAttr(nil, nil, slog.Group("named", slog.String("k", "v")))
	require.Len(t, namedGroup, 1)
	assert.Equal(t, "named.k", namedGroup[0].Key())
}

func TestSlogHandlerZeroTime(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(hqgologgerformatter.NewJSON(nil)), WithCallerCapture(true))

	record := slog.NewRecord(time.Time{}, slog.LevelInfo, "hello", 0)

	require.NoError(t, l.SlogHandler().Handle(context.Background(), record))

	decoded := decodeRecord(t, sink)
	assert.Contains(t, decoded, "time")
	assert.Contains(t, decoded["caller"], "slog_test.go:")
}
