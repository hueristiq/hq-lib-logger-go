package logger

import (
	"github.com/hueristiq/hq-lib-logger-go/formatter"
)

// Hook inspects a record before formatting. Hooks run while the logger holds
// its read lock and must not call logger configuration methods. Errors
// returned by Fire are discarded.
//
// Methods:
//   - Fire(log *formatter.Log) (err error): inspects a record before formatting.
type Hook interface {
	Fire(log *formatter.Log) (err error)
}

// HookFunc adapts a function to Hook. A nil function is a no-op.
type HookFunc func(log *formatter.Log) (err error)

// Fire calls f(log), or does nothing when f is nil.
//
// Parameters:
//   - log (*formatter.Log): record to inspect.
//
// Returns:
//   - err (error): the function result, or nil when f is nil.
func (f HookFunc) Fire(log *formatter.Log) (err error) {
	if f == nil {
		return nil
	}

	return f(log)
}

var _ Hook = HookFunc(nil)
