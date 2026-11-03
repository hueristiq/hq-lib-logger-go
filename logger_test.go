package logger_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

func newTestLogger(t *testing.T, buf *bytes.Buffer, level hqgologgerlevel.Level) *hqgologger.Logger {
	t.Helper()

	return hqgologger.New(
		hqgologger.WithLevel(level),
		hqgologger.WithFormatter(hqgologgerformatter.NewJSON(nil)),
		hqgologger.WithWriter(hqgologgerwriter.NewStream(buf)),
	)
}

func TestNewAndFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := newTestLogger(t, &buf, hqgologgerlevel.LevelInfo)

	logger.Info("hello",
		hqgologger.String("k", "v"),
		hqgologger.Int("n", 7),
		hqgologger.Bool("ok", true),
	)

	output := buf.String()

	assert.Contains(t, output, `"message":"hello"`)
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"k":"v"`)
	assert.Contains(t, output, `"n":7`)
	assert.Contains(t, output, `"ok":true`)
}

func TestLevelFilteringAndSetLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := newTestLogger(t, &buf, hqgologgerlevel.LevelDebug)

	assert.True(t, logger.Enabled(hqgologgerlevel.LevelDebug))
	assert.False(t, logger.Enabled(hqgologgerlevel.LevelTrace))
	assert.Equal(t, hqgologgerlevel.LevelDebug, logger.Level())

	logger.Trace("hidden")
	assert.NotContains(t, buf.String(), "hidden")

	require.NoError(t, logger.SetLevel(hqgologgerlevel.LevelTrace))
	assert.True(t, logger.Enabled(hqgologgerlevel.LevelTrace))

	logger.Trace("visible")
	assert.Contains(t, buf.String(), "visible")
}

func TestNamedAndPrefixViews(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := newTestLogger(t, &buf, hqgologgerlevel.LevelInfo)

	child := logger.Named("api").With(hqgologger.String("k", "v")).WithPrefix("pre:")
	child.Info("msg")

	output := buf.String()

	assert.Contains(t, output, `"logger":"api"`)
	assert.Contains(t, output, `"k":"v"`)
	assert.Contains(t, output, `"message":"pre:msg"`)

	// The parent view is unchanged.
	buf.Reset()
	logger.Info("parent")
	assert.NotContains(t, buf.String(), `"logger":"api"`)
	assert.NotContains(t, buf.String(), "pre:")
}

func TestSyncAndClose(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := newTestLogger(t, &buf, hqgologgerlevel.LevelInfo)

	require.NoError(t, logger.Sync())
	require.NoError(t, logger.Close())
	require.NoError(t, logger.Close())
}

func ExampleNew() {
	var buffer bytes.Buffer

	l := hqgologger.New(
		hqgologger.WithWriter(hqgologgerwriter.NewStream(&buffer)),
		hqgologger.WithFormatter(hqgologgerformatter.NewConsole(&hqgologgerformatter.ConsoleConfig{Timestamp: false, Label: true})),
	)

	l.Info("hello", hqgologger.String("service", "api"))

	fmt.Print(buffer.String())
	// Output:
	// [INF] hello service=api
}

func ExampleLogger_Infof() {
	var buffer bytes.Buffer

	l := hqgologger.New(
		hqgologger.WithWriter(hqgologgerwriter.NewStream(&buffer)),
		hqgologger.WithFormatter(hqgologgerformatter.NewConsole(&hqgologgerformatter.ConsoleConfig{Timestamp: false, Label: true})),
	)

	l.Infof("hello %s", "world")

	fmt.Print(buffer.String())
	// Output:
	// [INF] hello world
}
