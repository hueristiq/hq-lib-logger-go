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

	require.NoError(t, HookFunc(nil).Fire(&formatter.Log{Message: "ignored"}))
}

func TestHookFuncFire(t *testing.T) {
	t.Parallel()

	log := &formatter.Log{Message: "message"}

	var received *formatter.Log

	err := HookFunc(func(l *formatter.Log) error {
		received = l

		return nil
	}).Fire(log)

	require.NoError(t, err)
	assert.Same(t, log, received)
}

func TestHookFuncFireError(t *testing.T) {
	t.Parallel()

	expected := errors.New("hook failed")

	require.ErrorIs(t, HookFunc(func(*formatter.Log) error { return expected }).Fire(&formatter.Log{}), expected)
}
