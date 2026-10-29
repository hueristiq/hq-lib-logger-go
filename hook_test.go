package logger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-logger-go/formatter"
)

func TestHookFuncNilIsNoOp(t *testing.T) {
	t.Parallel()

	require.NoError(t, HookFunc(nil).Fire(formatter.Log{Message: "ignored"}))
}

func TestHookFuncFire(t *testing.T) {
	t.Parallel()

	log := formatter.Log{Message: "message"}

	var received formatter.Log

	err := HookFunc(func(l formatter.Log) error {
		received = l

		return nil
	}).Fire(log)

	require.NoError(t, err)
	assert.Equal(t, log, received)
}

func TestHookFuncFireError(t *testing.T) {
	t.Parallel()

	expected := errors.New("hook failed")

	require.ErrorIs(t, HookFunc(func(formatter.Log) error { return expected }).Fire(formatter.Log{}), expected)
}

type levelFilterHook struct {
	fired  []Level
	levels []Level
}

func (h *levelFilterHook) Levels() []Level { return h.levels }

//nolint:gocritic // Matches the Hook interface signature.
func (h *levelFilterHook) Fire(log formatter.Log) error {
	h.fired = append(h.fired, log.Level)

	return nil
}

func TestLevelFilteredHook(t *testing.T) {
	t.Parallel()

	sink := &captureWriter{}
	hook := &levelFilterHook{levels: []Level{LevelError}}
	l := New(WithWriter(sink), WithHooks(hook), WithLevel(LevelTrace))

	l.Info("info")
	l.Error("error")

	require.Equal(t, []Level{LevelError}, hook.fired)

	hook.levels = nil

	l.Error("error again")

	assert.Equal(t, []Level{LevelError}, hook.fired)
}
