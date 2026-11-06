package formatter_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

func TestConsoleFormatter(t *testing.T) {
	t.Parallel()

	data, err := hqgologgerformatter.NewConsole(nil).Format(testLog())

	require.NoError(t, err)
	assert.Contains(t, string(data), "msg")
	assert.Contains(t, string(data), "k=v")
}

func TestConsoleStripsControlCharacters(t *testing.T) {
	t.Parallel()

	log := hqgologgerformatter.Log{
		Time:    time.Unix(0, 0).UTC(),
		Level:   hqgologgerlevel.LevelInfo,
		Message: "hello\r\n\x1b[31mred\x07",
		Name:    "na\x1bme",
		Caller:  "cal\rler",
		Fields: []hqgologgerformatter.Field{
			hqgologgerformatter.String("ke\ty", "value\n\x1b[0m"),
		},
	}

	data, err := hqgologgerformatter.NewConsole(nil).Format(log)
	require.NoError(t, err)

	output := string(data)

	for _, control := range []string{"\x1b", "\r", "\x07", "\n"} {
		assert.NotContains(t, output, control)
	}

	assert.Contains(t, output, "hello")
	assert.Contains(t, output, "red")
	assert.Contains(t, output, "value")
	assert.Contains(t, output, "ke\ty")
}
