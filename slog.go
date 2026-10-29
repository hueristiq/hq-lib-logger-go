package logger

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/hueristiq/hq-lib-logger-go/formatter"
	"github.com/hueristiq/hq-lib-logger-go/levels"
)

// slogHandler adapts a Logger to slog.Handler.
//
// Fields:
//   - logger (*Logger): target logger.
//   - attrs ([]Field): fields attached by WithAttrs.
//   - groups ([]string): active group path.
type slogHandler struct {
	logger *Logger
	attrs  []Field
	groups []string
}

// NewSlogHandler adapts l to slog.Handler. A nil logger uses [New].
//
// Parameters:
//   - l (*Logger): target logger; nil uses a default logger.
//
// Returns:
//   - handler (slog.Handler): the adapter.
func NewSlogHandler(l *Logger) (handler slog.Handler) {
	if l == nil {
		l = New()
	}

	return &slogHandler{logger: l}
}

// SlogHandler adapts l to slog.Handler.
//
// Returns:
//   - handler (slog.Handler): the adapter.
func (l *Logger) SlogHandler() (handler slog.Handler) {
	return NewSlogHandler(l)
}

// Enabled reports whether the handler would emit records at the given slog
// level, mapped through [slogToLevel].
//
// Parameters:
//   - ctx (context.Context): request context; ignored.
//   - level (slog.Level): slog severity to test.
//
// Returns:
//   - enabled (bool): true when the level passes the logger threshold.
func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.logger.Enabled(slogToLevel(level))
}

// Handle formats one slog record through the logger and returns nil. The
// record's caller is honored when caller capture is enabled, with a
// captureCaller fallback.
//
// Parameters:
//   - ctx (context.Context): request context; ignored.
//   - record (slog.Record): record to emit.
//
// Returns:
//   - err (error): always nil.
//
//nolint:gocritic // slog.Handler requires passing slog.Record by value.
func (h *slogHandler) Handle(_ context.Context, record slog.Record) error {
	fields := slices.Clone(h.attrs)

	record.Attrs(func(attr slog.Attr) bool {
		fields = appendSlogAttr(fields, h.groups, attr)

		return true
	})

	timestamp := record.Time
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	caller := ""

	if h.logger.getCore().reportCaller.Load() {
		caller = callerFromPC(record.PC)

		if caller == "" {
			caller = captureCaller()
		}
	}

	h.logger.LogRecord(formatter.Log{
		Time:    timestamp,
		Level:   slogToLevel(record.Level),
		Message: record.Message,
		Caller:  caller,
		Fields:  fields,
	})

	return nil
}

// WithAttrs returns a handler copy with the given attributes converted into
// attached fields.
//
// Parameters:
//   - attrs ([]slog.Attr): attributes to attach.
//
// Returns:
//   - handler (slog.Handler): the derived handler.
func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := slices.Clone(h.attrs)

	for _, attr := range attrs {
		fields = appendSlogAttr(fields, h.groups, attr)
	}

	return &slogHandler{
		logger: h.logger,
		attrs:  fields,
		groups: slices.Clone(h.groups),
	}
}

// WithGroup returns a handler copy with an added group name for subsequent
// attributes. An empty name returns the handler unchanged.
//
// Parameters:
//   - name (string): group name.
//
// Returns:
//   - handler (slog.Handler): the derived handler.
func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	groups := make([]string, 0, len(h.groups)+1)
	groups = append(groups, h.groups...)
	groups = append(groups, name)

	return &slogHandler{
		logger: h.logger,
		attrs:  slices.Clone(h.attrs),
		groups: groups,
	}
}

// slogToLevel maps a slog level to the nearest package level: values below
// slog's Debug map to Trace, values at or above Error map to Error.
//
// Parameters:
//   - level (slog.Level): slog severity to map.
//
// Returns:
//   - mapped (levels.Level): the mapped package level.
func slogToLevel(level slog.Level) (mapped levels.Level) {
	switch {
	case level < slog.LevelDebug:
		return levels.LevelTrace
	case level < slog.LevelInfo:
		return levels.LevelDebug
	case level < slog.LevelWarn:
		return levels.LevelInfo
	case level < slog.LevelError:
		return levels.LevelWarn
	default:
		return levels.LevelError
	}
}

// appendSlogAttr converts a slog attribute into fields, prefixing the key
// with the active group path and flattening inline groups.
//
// Parameters:
//   - fields ([]Field): destination fields.
//   - groups ([]string): active group path.
//   - attr (slog.Attr): attribute to convert.
//
// Returns:
//   - out ([]Field): fields with the attribute appended.
func appendSlogAttr(fields []Field, groups []string, attr slog.Attr) []Field {
	value := attr.Value.Resolve()
	key := attr.Key

	if value.Kind() == slog.KindGroup {
		childGroups := groups

		if key != "" {
			childGroups = append(slices.Clone(groups), key)
		}

		for _, child := range value.Group() {
			fields = appendSlogAttr(fields, childGroups, child)
		}

		return fields
	}

	if key == "" {
		return fields
	}

	if len(groups) > 0 {
		key = joinGroup(groups, key)
	}

	switch value.Kind() {
	case slog.KindString:
		return append(fields, String(key, value.String()))
	case slog.KindInt64:
		return append(fields, Int64(key, value.Int64()))
	case slog.KindUint64:
		return append(fields, Uint64(key, value.Uint64()))
	case slog.KindFloat64:
		return append(fields, Float64(key, value.Float64()))
	case slog.KindBool:
		return append(fields, Bool(key, value.Bool()))
	case slog.KindDuration:
		return append(fields, Duration(key, value.Duration()))
	case slog.KindTime:
		return append(fields, Time(key, value.Time()))
	case slog.KindGroup:
		// Unreachable: groups are handled above. Kept for exhaustiveness.
		return fields
	case slog.KindAny, slog.KindLogValuer:
		// KindLogValuer is unreachable here: Resolve fully unwinds LogValuer
		// chains in current slog versions. Kept for forward compatibility.
		return append(fields, Any(key, value.Any()))
	default:
		// Unreachable: all slog kinds are listed above. Kept defensively.
		return append(fields, Any(key, value.Any()))
	}
}

// joinGroup joins the active group path with a key using a dot separator.
//
// Parameters:
//   - groups ([]string): active group path.
//   - key (string): attribute key.
//
// Returns:
//   - joined (string): the dotted full key.
func joinGroup(groups []string, key string) string {
	return strings.Join(groups, ".") + "." + key
}

var _ slog.Handler = (*slogHandler)(nil)
