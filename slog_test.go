package logger

import (
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-logger-go/formatter"
)

func decodeRecord(t *testing.T, sink *captureWriter) (decoded map[string]any) {
	t.Helper()

	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(sink.String())), &decoded))

	return decoded
}

func TestSlogHandlerBasic(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(formatter.NewJSON(nil)))

	slog.New(l.SlogHandler()).Info("hello", "request_id", "42")

	decoded := decodeRecord(t, sink)

	assert.Equal(t, "info", decoded["level"])
	assert.Equal(t, "hello", decoded["message"])
	assert.Equal(t, "42", decoded["request_id"])
}

func TestSlogHandlerAttrsAndGroups(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	l := New(WithWriter(sink), WithFormatter(formatter.NewJSON(nil)))

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
	l := New(WithWriter(sink), WithFormatter(formatter.NewJSON(nil)))

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
		WithFormatter(formatter.NewJSON(nil)),
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
