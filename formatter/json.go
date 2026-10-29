package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hueristiq/hq-lib-logger-go/levels"
)

// JSONConfig controls the JSON formatter. A nil config uses the package
// defaults, including logger, caller, and error keys. With a non-nil config,
// empty optional keys omit that part of the record.
//
// Fields:
//   - Timestamp (bool): include the record timestamp.
//   - TimestampFormat (string): timestamp layout; empty means RFC3339Nano.
//   - TimestampKey (string): timestamp key; empty means "time".
//   - LevelKey (string): level key; empty means "level".
//   - MessageKey (string): message key; empty means "message".
//   - NameKey (string): logger name key; empty omits the name.
//   - CallerKey (string): caller key; empty omits the caller.
//   - ErrorKey (string): error key; empty renders errors as ordinary fields.
//   - FieldsKey (string): nested fields object key; empty puts fields at top level.
//   - Indent (string): pretty-print indentation; empty is compact output.
//   - PreserveFieldOrder (bool): preserve field order instead of sorting by key.
type JSONConfig struct {
	Timestamp          bool
	TimestampFormat    string
	TimestampKey       string
	LevelKey           string
	MessageKey         string
	NameKey            string
	CallerKey          string
	ErrorKey           string
	FieldsKey          string
	Indent             string
	PreserveFieldOrder bool
}

// JSON emits one JSON object per record. The zero value is a usable compact
// formatter; use [NewJSON] for defaults.
//
// Fields:
//   - cfg (JSONConfig): formatter configuration.
type JSON struct {
	cfg JSONConfig
}

// NewJSON returns a JSON formatter. A nil config uses the defaults.
//
// Parameters:
//   - cfg (*JSONConfig): formatter configuration; nil uses defaults.
//
// Returns:
//   - formatter (*JSON): the JSON formatter.
func NewJSON(cfg *JSONConfig) (formatter *JSON) {
	if cfg == nil {
		cfg = &JSONConfig{
			Timestamp:       true,
			TimestampFormat: time.RFC3339Nano,
			TimestampKey:    "time",
			LevelKey:        "level",
			MessageKey:      "message",
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
		copied.MessageKey = "message"
	}

	return &JSON{cfg: copied}
}

// Format renders log.
//
// Parameters:
//   - log (*Log): record to format; must be non-nil.
//
// Returns:
//   - data ([]byte): the JSON record.
//   - err (error): [ErrNilLog] when log is nil, or an invalid-level error.
func (j *JSON) Format(log *Log) (data []byte, err error) {
	if log == nil {
		return nil, ErrNilLog
	}

	return j.Append(nil, *log)
}

// Append appends the JSON representation of log to dst.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//   - log (Log): record to format.
//
// Returns:
//   - data ([]byte): dst with the JSON record appended.
//   - err (error): non-nil when the level is invalid or indentation fails.
//
//nolint:gocritic // Log is passed by value to match the Appender interface.
func (j *JSON) Append(dst []byte, log Log) (data []byte, err error) {
	if !log.Level.Valid() {
		return dst, fmt.Errorf("%w (%d)", levels.ErrUnknownLevel, int(log.Level))
	}

	fields := log.Fields
	if !j.cfg.PreserveFieldOrder && len(fields) > 1 {
		fields = slices.Clone(fields)
		slices.SortStableFunc(fields, func(a, b Field) int {
			return strings.Compare(a.key, b.key)
		})
	}

	var errorText string

	if j.cfg.ErrorKey != "" {
		for i := range fields {
			if fields[i].key != j.cfg.ErrorKey || fields[i].isNilAny() {
				continue
			}

			errorText = string(fields[i].AppendText(nil))
		}
	}

	objectStart := len(dst)

	dst = append(dst, '{')

	first := true
	writeKey := func(key string) {
		if !first {
			dst = append(dst, ',')
		}

		first = false
		dst = appendJSONString(dst, key)
		dst = append(dst, ':')
	}

	if j.cfg.Timestamp && !log.Time.IsZero() {
		writeKey(j.cfg.TimestampKey)
		dst = appendJSONString(dst, log.Time.Format(j.cfg.TimestampFormat))
	}

	writeKey(j.cfg.LevelKey)

	dst = appendJSONString(dst, log.Level.String())

	writeKey(j.cfg.MessageKey)

	dst = appendJSONString(dst, log.Message)

	if j.cfg.NameKey != "" && log.Name != "" {
		writeKey(j.cfg.NameKey)

		dst = appendJSONString(dst, log.Name)
	}

	if j.cfg.CallerKey != "" && log.Caller != "" {
		writeKey(j.cfg.CallerKey)

		dst = appendJSONString(dst, log.Caller)
	}

	if errorText != "" {
		writeKey(j.cfg.ErrorKey)

		dst = appendJSONString(dst, errorText)
	}

	timestampWritten := j.cfg.Timestamp && !log.Time.IsZero()
	nameWritten := j.cfg.NameKey != "" && log.Name != ""
	callerWritten := j.cfg.CallerKey != "" && log.Caller != ""
	errorWritten := errorText != ""

	writeField := func(field *Field) {
		reserved := field.key == "" ||
			(field.key == j.cfg.TimestampKey && timestampWritten) ||
			field.key == j.cfg.LevelKey || field.key == j.cfg.MessageKey ||
			(field.key == j.cfg.NameKey && nameWritten) ||
			(field.key == j.cfg.CallerKey && callerWritten) ||
			(field.key == j.cfg.ErrorKey && errorWritten) ||
			(field.key == j.cfg.FieldsKey && j.cfg.FieldsKey != "")
		if reserved || field.isNilAny() {
			return
		}

		writeKey(field.key)

		if field.kind == kindAny {
			if errValue, ok := field.any.(error); ok {
				dst = appendJSONString(dst, nilSafeErrorText(errValue))

				return
			}
		}

		written, writeErr := field.AppendJSON(dst)
		if writeErr != nil {
			dst = appendJSONString(dst, string(field.AppendText(nil)))

			return
		}

		dst = written
	}

	if j.cfg.FieldsKey != "" {
		writeKey(j.cfg.FieldsKey)

		dst = append(dst, '{')

		fieldsFirst := first
		first = true

		for i := range fields {
			writeField(&fields[i])
		}

		first = fieldsFirst

		dst = append(dst, '}')
	} else {
		for i := range fields {
			writeField(&fields[i])
		}
	}

	dst = append(dst, '}')

	if j.cfg.Indent != "" {
		var indented bytes.Buffer

		if indentErr := json.Indent(&indented, dst[objectStart:], "", j.cfg.Indent); indentErr != nil {
			return dst, indentErr
		}

		dst = append(dst[:objectStart], indented.Bytes()...)
	}

	return dst, nil
}

var (
	_ Formatter = (*JSON)(nil)
	_ Appender  = (*JSON)(nil)
)
