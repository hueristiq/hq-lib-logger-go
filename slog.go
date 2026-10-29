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

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.logger.Enabled(slogToLevel(level))
}

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
		return fields
	case slog.KindAny, slog.KindLogValuer:
		return append(fields, Any(key, value.Any()))
	default:
		return append(fields, Any(key, value.Any()))
	}
}

func joinGroup(groups []string, key string) string {
	return strings.Join(groups, ".") + "." + key
}

var _ slog.Handler = (*slogHandler)(nil)
