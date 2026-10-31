package formatter_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

func testLog() hqgologgerformatter.Log {
	return hqgologgerformatter.Log{
		Time:    time.Unix(0, 0).UTC(),
		Level:   hqgologgerlevel.LevelInfo,
		Message: "msg",
		Fields: []hqgologgerformatter.Field{
			hqgologgerformatter.String("k", "v"),
			hqgologgerformatter.Int("n", 7),
		},
	}
}

func TestJSONFormatter(t *testing.T) {
	t.Parallel()

	data, err := hqgologgerformatter.NewJSON(nil).Format(testLog())

	require.NoError(t, err)
	assert.True(t, json.Valid(data))
	assert.Contains(t, string(data), `"message":"msg"`)
	assert.Contains(t, string(data), `"level":"info"`)
	assert.Contains(t, string(data), `"k":"v"`)
	assert.Contains(t, string(data), `"n":7`)
}

func TestLogfmtFormatter(t *testing.T) {
	t.Parallel()

	data, err := hqgologgerformatter.NewLogfmt(nil).Format(testLog())

	require.NoError(t, err)
	assert.Contains(t, string(data), "msg")
	assert.Contains(t, string(data), "k=v")
}

func TestConsoleFormatter(t *testing.T) {
	t.Parallel()

	data, err := hqgologgerformatter.NewConsole(nil).Format(testLog())

	require.NoError(t, err)
	assert.Contains(t, string(data), "msg")
	assert.Contains(t, string(data), "k=v")
}

func TestFormatterRejectsInvalidLevel(t *testing.T) {
	t.Parallel()

	log := testLog()
	log.Level = hqgologgerlevel.Level(99)

	_, err := hqgologgerformatter.NewJSON(nil).Format(log)

	require.Error(t, err)
	assert.ErrorIs(t, err, hqgologgerlevel.ErrUnknownLevel)
}
