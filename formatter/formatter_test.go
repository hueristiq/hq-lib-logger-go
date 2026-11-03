package formatter_test

import (
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

func TestFormatterRejectsInvalidLevel(t *testing.T) {
	t.Parallel()

	log := testLog()
	log.Level = hqgologgerlevel.Level(99)

	_, err := hqgologgerformatter.NewJSON(nil).Format(log)

	require.Error(t, err)
	assert.ErrorIs(t, err, hqgologgerlevel.ErrUnknownLevel)
}
