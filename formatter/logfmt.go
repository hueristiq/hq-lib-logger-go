package formatter

import (
	"fmt"
	"slices"
	"strconv"
	"time"
	"unicode/utf8"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
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
//   - log (Log): record to format.
//
// Returns:
//   - data ([]byte): the logfmt record.
//   - err (error): non-nil when the level is invalid.
//
//nolint:gocritic // Log is passed by value to match the Appender interface and avoid an escape.
func (l *Logfmt) Format(log Log) (data []byte, err error) {
	return l.Append(nil, log)
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
		return dst, fmt.Errorf("%w (%d)", hqgologgerlevel.ErrUnknownLevel, int(log.Level))
	}

	fields := log.Fields

	var indices *[]int

	if !l.cfg.PreserveFieldOrder && len(fields) > 1 {
		indices = acquireFieldIndices(len(fields))
		defer releaseFieldIndices(indices)

		perm := *indices
		for i := range perm {
			perm[i] = i
		}

		slices.SortStableFunc(perm, func(a, b int) int {
			switch {
			case fields[a].key < fields[b].key:
				return -1
			case fields[a].key > fields[b].key:
				return 1
			default:
				return 0
			}
		})
	}

	fieldAt := func(i int) *Field {
		if indices != nil {
			return &fields[(*indices)[i]]
		}

		return &fields[i]
	}

	var errorText string

	if l.cfg.ErrorKey != "" {
		for i := range fields {
			field := fieldAt(i)

			if field.key != l.cfg.ErrorKey || field.isNilAny() {
				continue
			}

			errorText = string(field.AppendText(nil))
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
		if logfmtSafe(l.cfg.TimestampFormat) {
			// The timestamp is always the first pair, so no separator is
			// needed before it.
			first = false
			dst = appendLogfmtKey(dst, l.cfg.TimestampKey)
			dst = append(dst, '=')
			dst = log.Time.AppendFormat(dst, l.cfg.TimestampFormat)
		} else {
			writePair(l.cfg.TimestampKey, log.Time.Format(l.cfg.TimestampFormat))
		}
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
		field := fieldAt(i)

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

// appendLogfmtKey appends a logfmt key, quoting it when it contains unsafe
// bytes.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//   - key (string): key to append.
//
// Returns:
//   - out ([]byte): dst with the key appended.
func appendLogfmtKey(dst []byte, key string) []byte {
	if key != "" && logfmtSafe(key) {
		return append(dst, key...)
	}

	return strconv.AppendQuote(dst, key)
}

// appendLogfmtFieldValue appends a field value in logfmt form, quoting
// string, error, and arbitrary values when they contain unsafe bytes.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//   - field (*Field): field to append.
//
// Returns:
//   - out ([]byte): dst with the value appended.
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

// appendLogfmtValue appends a logfmt value, quoting it when it is empty or
// contains unsafe bytes.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//   - value (string): value to append.
//
// Returns:
//   - out ([]byte): dst with the value appended.
func appendLogfmtValue(dst []byte, value string) []byte {
	if value == "" || logfmtSafe(value) {
		return append(dst, value...)
	}

	return strconv.AppendQuote(dst, value)
}

// logfmtSafe reports whether s needs no quoting in logfmt: valid UTF-8 with
// no whitespace, control characters, quotes, equals signs, or backslashes.
//
// Parameters:
//   - s (string): candidate value.
//
// Returns:
//   - safe (bool): true when s needs no quoting.
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
