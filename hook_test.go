package logger_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

func TestHookFunc(t *testing.T) {
	t.Parallel()

	var received hqgologgerformatter.Log

	var hook hqgologger.HookFunc = func(log hqgologgerformatter.Log) error {
		received = log

		return nil
	}

	require.NoError(t, hook.Fire(hqgologgerformatter.Log{Message: "msg"}))
	assert.Equal(t, "msg", received.Message)

	var nilHook hqgologger.HookFunc

	assert.NoError(t, nilHook.Fire(hqgologgerformatter.Log{}))
}

func TestWithHooksFiresBeforeFormatting(t *testing.T) {
	t.Parallel()

	var (
		buf    bytes.Buffer
		fired  int
		level  hqgologgerlevel.Level
		fields int
	)

	hook := hqgologger.HookFunc(func(log hqgologgerformatter.Log) error {
		fired++
		level = log.Level
		fields = len(log.Fields)

		return nil
	})

	logger := hqgologger.New(
		hqgologger.WithLevel(hqgologgerlevel.LevelInfo),
		hqgologger.WithFormatter(hqgologgerformatter.NewJSON(nil)),
		hqgologger.WithWriter(hqgologgerwriter.NewStream(&buf)),
		hqgologger.WithHooks(hook),
	)

	logger.Info("msg", hqgologger.String("k", "v"))

	assert.Equal(t, 1, fired)
	assert.Equal(t, hqgologgerlevel.LevelInfo, level)
	assert.Equal(t, 1, fields)
	assert.Contains(t, buf.String(), `"message":"msg"`)
}

func TestNilWriterLogger(t *testing.T) {
	t.Parallel()

	logger := hqgologger.New(
		hqgologger.WithFormatter(hqgologgerformatter.NewJSON(nil)),
		hqgologger.WithWriter(hqgologgerwriter.NewStream(io.Discard)),
	)

	logger.Info("msg")

	assert.True(t, logger.Enabled(hqgologgerlevel.LevelInfo))
}
