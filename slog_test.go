package logger_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

func TestSlogHandler(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := hqgologger.New(
		hqgologger.WithLevel(hqgologgerlevel.LevelDebug),
		hqgologger.WithFormatter(hqgologgerformatter.NewJSON(nil)),
		hqgologger.WithWriter(hqgologgerwriter.NewStream(&buf)),
	)

	handler := logger.SlogHandler()

	require.NotNil(t, handler)
	assert.True(t, handler.Enabled(context.Background(), slog.LevelInfo))

	slog.New(handler).Info("from slog", "k", "v")

	output := buf.String()
	assert.Contains(t, output, `"message":"from slog"`)
	assert.Contains(t, output, `"k":"v"`)
}

func TestSlogHandlerMethods(t *testing.T) {
	t.Parallel()

	logger := hqgologger.New()
	handler := logger.SlogHandler()

	require.NotNil(t, handler)

	withAttrs := handler.WithAttrs([]slog.Attr{slog.String("k", "v")})
	require.NotNil(t, withAttrs)

	withGroup := handler.WithGroup("group")
	require.NotNil(t, withGroup)

	var nilLogger *hqgologger.Logger

	assert.NotNil(t, nilLogger.SlogHandler())
}
