package levels

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelInt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, LevelFatal.Int())
	assert.Equal(t, 1, LevelSilent.Int())
	assert.Equal(t, 2, LevelError.Int())
	assert.Equal(t, 3, LevelInfo.Int())
	assert.Equal(t, 4, LevelWarn.Int())
	assert.Equal(t, 5, LevelDebug.Int())
}

func TestLevelString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level Level
		want  string
	}{
		{LevelFatal, "fatal"},
		{LevelSilent, "silent"},
		{LevelError, "error"},
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelDebug, "debug"},
		{Level(-1), "unknown"},
		{Level(99), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.level.String())
	}
}

func TestLevelIsValid(t *testing.T) {
	t.Parallel()

	assert.True(t, LevelFatal.IsValid())
	assert.True(t, LevelDebug.IsValid())
	assert.False(t, Level(-1).IsValid())
	assert.False(t, Level(6).IsValid())
}

func TestLevelMarshalText(t *testing.T) {
	t.Parallel()

	bytes, err := LevelInfo.MarshalText()

	require.NoError(t, err)
	assert.Equal(t, "info", string(bytes))
}

func TestLevelUnmarshalText(t *testing.T) {
	t.Parallel()

	var l Level

	require.NoError(t, l.UnmarshalText([]byte("warn")))
	assert.Equal(t, LevelWarn, l)
}

func TestLevelUnmarshalTextUnknown(t *testing.T) {
	t.Parallel()

	var l Level

	err := l.UnmarshalText([]byte("bogus"))

	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnknownLevel)
	assert.Contains(t, err.Error(), "bogus")
	assert.NotContains(t, err.Error(), "fatal")
}

func TestLevelRoundTrip(t *testing.T) {
	t.Parallel()

	for _, want := range []Level{
		LevelFatal,
		LevelSilent,
		LevelError,
		LevelInfo,
		LevelWarn,
		LevelDebug,
	} {
		text, err := want.MarshalText()
		require.NoError(t, err)

		var got Level

		require.NoError(t, got.UnmarshalText(text))
		assert.Equal(t, want, got)
	}
}

func TestLevelJSON(t *testing.T) {
	t.Parallel()

	type payload struct {
		Level Level `json:"level"`
	}

	data, err := json.Marshal(payload{Level: LevelError})
	require.NoError(t, err)
	assert.JSONEq(t, `{"level":"error"}`, string(data))

	var got payload

	require.NoError(t, json.Unmarshal([]byte(`{"level":"debug"}`), &got))
	assert.Equal(t, LevelDebug, got.Level)
}

func TestLevelUnmarshalTextAllNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		want Level
	}{
		{"fatal", LevelFatal},
		{"silent", LevelSilent},
		{"error", LevelError},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"debug", LevelDebug},
	}

	for _, tt := range tests {
		var l Level

		require.NoError(t, l.UnmarshalText([]byte(tt.text)))
		assert.Equal(t, tt.want, l)
	}
}

func TestLevelUnmarshalTextCaseSensitive(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"WARN", "Warn", " warn", "warn "} {
		var l Level

		err := l.UnmarshalText([]byte(text))

		require.ErrorIs(t, err, ErrUnknownLevel, "%q must not parse", text)
	}
}

func TestLevelUnmarshalTextEmpty(t *testing.T) {
	t.Parallel()

	var l Level

	err := l.UnmarshalText(nil)
	require.ErrorIs(t, err, ErrUnknownLevel)

	err = l.UnmarshalText([]byte{})
	require.ErrorIs(t, err, ErrUnknownLevel)
}

func TestLevelUnmarshalTextErrorLeavesLevelUnchanged(t *testing.T) {
	t.Parallel()

	l := LevelInfo

	require.ErrorIs(t, l.UnmarshalText([]byte("bogus")), ErrUnknownLevel)
	assert.Equal(t, LevelInfo, l)
}

func TestLevelMarshalTextInvalidLevel(t *testing.T) {
	t.Parallel()

	for _, level := range []Level{Level(-1), Level(99)} {
		text, err := level.MarshalText()

		require.NoError(t, err)
		assert.Equal(t, "unknown", string(text))
	}
}

func TestLevelJSONUnmarshalError(t *testing.T) {
	t.Parallel()

	type payload struct {
		Level Level `json:"level"`
	}

	var got payload

	err := json.Unmarshal([]byte(`{"level":"bogus"}`), &got)

	require.ErrorIs(t, err, ErrUnknownLevel)
}
