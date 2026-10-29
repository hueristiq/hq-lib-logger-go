package logger

import (
	"context"
)

// contextKey is the unexported key type for loggers stored in contexts. The
// private type prevents collisions with values stored by other packages.
type contextKey struct{}

// NewContext returns a context carrying l. A nil logger leaves ctx unchanged.
//
// Parameters:
//   - ctx (context.Context): parent context.
//   - l (*Logger): logger to store; nil is ignored.
//
// Returns:
//   - ctx (context.Context): the derived context.
func NewContext(ctx context.Context, l *Logger) (out context.Context) {
	if l == nil {
		return ctx
	}

	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext returns the logger stored by [NewContext].
//
// Parameters:
//   - ctx (context.Context): context to inspect.
//
// Returns:
//   - logger (*Logger): the stored logger, or nil.
//   - ok (bool): true when a logger was stored.
func FromContext(ctx context.Context) (logger *Logger, ok bool) {
	if ctx == nil {
		return nil, false
	}

	logger, ok = ctx.Value(contextKey{}).(*Logger)

	return logger, ok && logger != nil
}
