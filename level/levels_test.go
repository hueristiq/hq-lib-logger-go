package level_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

func TestLevelNamesAndValidity(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "fatal", hqgologgerlevel.LevelFatal.String())
	assert.Equal(t, "error", hqgologgerlevel.LevelError.String())
	assert.Equal(t, "warn", hqgologgerlevel.LevelWarn.String())
	assert.Equal(t, "info", hqgologgerlevel.LevelInfo.String())
	assert.Equal(t, "debug", hqgologgerlevel.LevelDebug.String())
	assert.Equal(t, "trace", hqgologgerlevel.LevelTrace.String())
	assert.Equal(t, "unknown", hqgologgerlevel.Level(99).String())

	assert.True(t, hqgologgerlevel.LevelInfo.Valid())
	assert.False(t, hqgologgerlevel.Level(99).Valid())
}

func TestParse(t *testing.T) {
	t.Parallel()

	parsed, err := hqgologgerlevel.Parse(" WARN ")

	require.NoError(t, err)
	assert.Equal(t, hqgologgerlevel.LevelWarn, parsed)

	parsed, err = hqgologgerlevel.Parse("DEBUG")

	require.NoError(t, err)
	assert.Equal(t, hqgologgerlevel.LevelDebug, parsed)

	_, err = hqgologgerlevel.Parse("nope")

	require.Error(t, err)
	assert.ErrorIs(t, err, hqgologgerlevel.ErrUnknownLevel)
}

func TestMarshalText(t *testing.T) {
	t.Parallel()

	text, err := hqgologgerlevel.LevelInfo.MarshalText()

	require.NoError(t, err)
	assert.Equal(t, "info", string(text))

	_, err = hqgologgerlevel.Level(99).MarshalText()

	require.Error(t, err)
	assert.ErrorIs(t, err, hqgologgerlevel.ErrUnknownLevel)
}

func ExampleParse() {
	level, err := hqgologgerlevel.Parse(" warn ")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(level)
	// Output:
	// warn
}
