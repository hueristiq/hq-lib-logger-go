package formatter

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hueristiq/hq-lib-logger-go/levels"
)

// LogfmtConfig controls the logfmt formatter. A nil config uses the package
// defaults, including logger, caller, and error keys. With a non-nil config,
// empty optional keys omit that part of the record.
//
// Fields:
//   - Timestamp (bool): include the record timestamp.
//   - TimestampFormat (string): timestamp layout; empty means RFC3339Nano.
//   - TimestampKey (string): timestamp key; empty means "time".
//   - LevelKey (string): level key; empty means "level".
//   - MessageKey (string): message key; empty means "msg".
//   - NameKey (string): logger name key; empty omits the name.
//   - CallerKey (string): caller key; empty omits the caller.
//   - ErrorKey (string): error key; empty renders errors as ordinary fields.
//   - PreserveFieldOrder (bool): preserve field order instead of sorting by key.
type LogfmtConfig struct {
	Timestamp          bool
	TimestampFormat    string
	TimestampKey       string
	LevelKey           string
	MessageKey         string
	NameKey            string
	CallerKey          string
	ErrorKey           string
	PreserveFieldOrder bool
}

// Logfmt emits compact key=value lines. The zero value is a usable formatter;
// use [NewLogfmt] for defaults.
//
// Fields:
//   - cfg (LogfmtConfig): formatter configuration.
type Logfmt struct {
	cfg LogfmtConfig
}

// NewLogfmt returns a logfmt formatter. A nil config uses the defaults.
//
// Parameters:
//   - cfg (*LogfmtConfig): formatter configuration; nil uses defaults.
//
// Returns:
//   - formatter (*Logfmt): the logfmt formatter.
func NewLogfmt(cfg *LogfmtConfig) (formatter *Logfmt) {
	if cfg == nil {
		cfg = &LogfmtConfig{
			Timestamp:       true,
			TimestampFormat: time.RFC3339Nano,
			TimestampKey:    "time",
			LevelKey:        "level",
			MessageKey:      "msg",
			NameKey:         "logger",
			CallerKey:       "caller",
			ErrorKey:        "error",
		}
	}

	copied := *cfg
	if copied.TimestampFormat == "" {
		copied.TimestampFormat = time.RFC3339Nano
	}

	if copied.TimestampKey == "" {
		copied.TimestampKey = "time"
	}

	if copied.LevelKey == "" {
		copied.LevelKey = "level"
	}

	if copied.MessageKey == "" {
		copied.MessageKey = "msg"
	}

	return &Logfmt{cfg: copied}
}

// Format renders log.
//
// Parameters:
//   - log (*Log): record to format; must be non-nil.
//
// Returns:
//   - data ([]byte): the logfmt record.
//   - err (error): [ErrNilLog] when log is nil, or an invalid-level error.
func (l *Logfmt) Format(log *Log) (data []byte, err error) {
	if log == nil {
		return nil, ErrNilLog
	}

	return l.Append(nil, *log)
}

// Append appends the logfmt representation of log to dst.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//   - log (Log): record to format.
//
// Returns:
//   - data ([]byte): dst with the logfmt record appended.
//   - err (error): non-nil when the level is invalid.
//
//nolint:gocritic // Log is passed by value to match the Appender interface.
func (l *Logfmt) Append(dst []byte, log Log) (data []byte, err error) {
	if !log.Level.Valid() {
		return dst, fmt.Errorf("%w (%d)", levels.ErrUnknownLevel, int(log.Level))
	}

	fields := log.Fields
	if !l.cfg.PreserveFieldOrder && len(fields) > 1 {
		fields = slices.Clone(fields)
		slices.SortStableFunc(fields, func(a, b Field) int {
			return strings.Compare(a.key, b.key)
		})
	}

	var errorText string

	if l.cfg.ErrorKey != "" {
		for i := range fields {
			if fields[i].key != l.cfg.ErrorKey || fields[i].isNilAny() {
				continue
			}

			errorText = string(fields[i].AppendText(nil))
		}
	}

	first := true
	writePair := func(key, value string) {
		if !first {
			dst = append(dst, ' ')
		}

		first = false
		dst = appendLogfmtKey(dst, key)
		dst = append(dst, '=')
		dst = appendLogfmtValue(dst, value)
	}

	if l.cfg.Timestamp && !log.Time.IsZero() {
		writePair(l.cfg.TimestampKey, log.Time.Format(l.cfg.TimestampFormat))
	}

	writePair(l.cfg.LevelKey, log.Level.String())
	writePair(l.cfg.MessageKey, log.Message)

	if l.cfg.NameKey != "" && log.Name != "" {
		writePair(l.cfg.NameKey, log.Name)
	}

	if l.cfg.CallerKey != "" && log.Caller != "" {
		writePair(l.cfg.CallerKey, log.Caller)
	}

	if errorText != "" {
		writePair(l.cfg.ErrorKey, errorText)
	}

	timestampWritten := l.cfg.Timestamp && !log.Time.IsZero()
	nameWritten := l.cfg.NameKey != "" && log.Name != ""
	callerWritten := l.cfg.CallerKey != "" && log.Caller != ""
	errorWritten := errorText != ""

	for i := range fields {
		field := &fields[i]

		reserved := field.key == "" ||
			(field.key == l.cfg.ErrorKey && errorWritten) ||
			(field.key == l.cfg.TimestampKey && timestampWritten) ||
			field.key == l.cfg.LevelKey || field.key == l.cfg.MessageKey ||
			(field.key == l.cfg.NameKey && nameWritten) ||
			(field.key == l.cfg.CallerKey && callerWritten)
		if reserved || field.isNilAny() {
			continue
		}

		if !first {
			dst = append(dst, ' ')
		}

		first = false
		dst = appendLogfmtKey(dst, field.key)
		dst = append(dst, '=')
		dst = appendLogfmtFieldValue(dst, field)
	}

	return dst, nil
}

func appendLogfmtKey(dst []byte, key string) []byte {
	if key != "" && logfmtSafe(key) {
		return append(dst, key...)
	}

	return strconv.AppendQuote(dst, key)
}

func appendLogfmtFieldValue(dst []byte, field *Field) []byte {
	switch field.kind { //nolint:exhaustive // Scalar kinds share AppendText.
	case kindString:
		return appendLogfmtValue(dst, field.str)
	case kindError, kindAny:
		return appendLogfmtValue(dst, nilSafeErrorText(field.any))
	default:
		return field.AppendText(dst)
	}
}

func appendLogfmtValue(dst []byte, value string) []byte {
	if value == "" || logfmtSafe(value) {
		return append(dst, value...)
	}

	return strconv.AppendQuote(dst, value)
}

func logfmtSafe(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}

	for i := range len(s) {
		switch c := s[i]; {
		case c <= ' ', c == 0x7f, c == '"', c == '=', c == '\\':
			return false
		}
	}

	return true
}

var (
	_ Formatter = (*Logfmt)(nil)
	_ Appender  = (*Logfmt)(nil)
)
