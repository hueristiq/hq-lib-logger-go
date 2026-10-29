package levels

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelStringAndValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level Level
		name  string
		valid bool
	}{
		{LevelFatal, "fatal", true},
		{LevelError, "error", true},
		{LevelWarn, "warn", true},
		{LevelInfo, "info", true},
		{LevelDebug, "debug", true},
		{LevelTrace, "trace", true},
		{Level(99), "unknown", false},
		{Level(-1), "unknown", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.name, tt.level.String())
		assert.Equal(t, tt.valid, tt.level.Valid())
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  Level
	}{
		{"fatal", LevelFatal},
		{"ERROR", LevelError},
		{"  Warn ", LevelWarn},
		{"info", LevelInfo},
		{"Debug", LevelDebug},
		{"trace", LevelTrace},
	}

	for _, tt := range tests {
		got, err := Parse(tt.input)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got)
	}

	got, err := Parse("verbose")
	require.ErrorIs(t, err, ErrUnknownLevel)
	assert.Equal(t, LevelFatal, got)

	_, err = Parse("")
	require.ErrorIs(t, err, ErrUnknownLevel)
}

func TestTextRoundTrip(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(struct {
		Level Level `json:"level"`
	}{Level: LevelDebug})
	require.NoError(t, err)
	assert.JSONEq(t, `{"level":"debug"}`, string(data))

	var decoded struct {
		Level Level `json:"level"`
	}

	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, LevelDebug, decoded.Level)

	_, err = Level(99).MarshalText()
	require.ErrorIs(t, err, ErrUnknownLevel)

	var invalid Level

	require.ErrorIs(t, invalid.UnmarshalText([]byte("verbose")), ErrUnknownLevel)
}

func ExampleParse() {
	level, err := Parse(" warn ")
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(level)
	// Output:
	// warn
}
