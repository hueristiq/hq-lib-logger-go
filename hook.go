package logger

import (
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

// Hook inspects a record before formatting. Hooks receive a copy of the
// record and run while the logger holds its read lock; they must not call
// logger configuration methods. Errors returned by Fire are discarded.
//
// Methods:
//   - Fire(log hqgologgerformatter.Log) (err error): inspects a copy of the record.
type Hook interface {
	Fire(log hqgologgerformatter.Log) (err error)
}

// HookFunc adapts a function to Hook. A nil function is a no-op.
type HookFunc func(log hqgologgerformatter.Log) (err error)

// LevelFilter is an optional Hook extension restricting the levels a hook
// fires for. A hook without Levels fires for every record; an empty list
// means the hook never fires.
//
// Methods:
//   - Levels() (levels []hqgologgerlevel.Level): the levels the hook applies to.
type LevelFilter interface {
	Levels() (levels []hqgologgerlevel.Level)
}

// Fire calls f(log), or does nothing when f is nil.
//
// Parameters:
//   - log (hqgologgerformatter.Log): record to inspect.
//
// Returns:
//   - err (error): the function result, or nil when f is nil.
//
//nolint:gocritic // Log is passed by value to match the Hook interface and avoid an escape.
func (f HookFunc) Fire(log hqgologgerformatter.Log) (err error) {
	if f == nil {
		return nil
	}

	return f(log)
}

var _ Hook = HookFunc(nil)
