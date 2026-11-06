package formatter

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/hueristiq/hq-lib-logger-go/internal/nilvalue"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

// ConsoleConfig controls the console formatter. A nil config uses the package
// defaults: timestamp and label enabled, [RFC 3339](https://www.rfc-editor.org/info/rfc3339/) timestamps, sorted fields.
//
// Fields:
//   - Timestamp (bool): include the record timestamp.
//   - TimestampFormat (string): timestamp layout; empty means [RFC 3339](https://www.rfc-editor.org/info/rfc3339/).
//   - Label (bool): include a bracketed label.
//   - Colorize (bool): colorize labels using Colorizer.
//   - Colorizer (Colorizer): label colorizer; nil becomes NoOpColorizer.
//   - PreserveFieldOrder (bool): preserve field order instead of sorting by key.
type ConsoleConfig struct {
	Timestamp          bool
	TimestampFormat    string
	Label              bool
	Colorize           bool
	Colorizer          Colorizer
	PreserveFieldOrder bool
}

// Console renders human-readable "[time] [label] [name] message fields" lines.
// The zero value is a usable plain formatter; use [NewConsole] for defaults.
//
// Fields:
//   - cfg (ConsoleConfig): formatter configuration.
type Console struct {
	cfg ConsoleConfig
}

// NewConsole returns a console formatter. A nil config uses the defaults.
// Messages, names, callers, and field text have control characters removed,
// so untrusted values cannot inject terminal escapes or extra lines.
//
// Parameters:
//   - cfg (*ConsoleConfig): formatter configuration; nil uses defaults.
//
// Returns:
//   - formatter (*Console): the console formatter.
func NewConsole(cfg *ConsoleConfig) (formatter *Console) {
	if cfg == nil {
		cfg = &ConsoleConfig{
			Timestamp:       true,
			TimestampFormat: time.RFC3339,
			Label:           true,
		}
	}

	copied := *cfg
	if copied.TimestampFormat == "" {
		copied.TimestampFormat = time.RFC3339
	}

	if copied.Colorize && nilvalue.IsNil(copied.Colorizer) {
		copied.Colorizer = NoOpColorizer{}
	}

	return &Console{cfg: copied}
}

// Format renders log.
//
// Parameters:
//   - log (Log): record to format.
//
// Returns:
//   - data ([]byte): the formatted record.
//   - err (error): non-nil when the level is invalid.
//
//nolint:gocritic // Log is passed by value to match the Appender interface and avoid an escape.
func (c *Console) Format(log Log) (data []byte, err error) {
	return c.Append(nil, log)
}

// Append appends the console representation of log to dst.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//   - log (Log): record to format.
//
// Returns:
//   - data ([]byte): dst with the formatted record appended.
//   - err (error): non-nil when the level is invalid.
//
//nolint:gocritic // Log is passed by value to match the Appender interface.
func (c *Console) Append(dst []byte, log Log) (data []byte, err error) {
	if !log.Level.Valid() {
		return dst, fmt.Errorf("%w (%d)", hqgologgerlevel.ErrUnknownLevel, int(log.Level))
	}

	if c.cfg.Timestamp && !log.Time.IsZero() {
		before := len(dst)

		dst = log.Time.AppendFormat(dst, c.cfg.TimestampFormat)
		if len(dst) > before {
			dst = append(dst, ' ')
		}
	}

	fields := log.Fields

	var indices *[]int

	if !c.cfg.PreserveFieldOrder && len(fields) > 1 {
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

	labelSet := false
	label := ""

	for i := range fields {
		field := fieldAt(i)

		if field.key != LabelKey {
			continue
		}

		labelSet = true

		if field.kind == kindString {
			label = field.str
		}

		break
	}

	if c.cfg.Label {
		if !labelSet {
			label = defaultLabels[int(log.Level)]
		}

		if label != "" {
			label = sanitizeConsoleText(label)
			colorized := label

			if c.cfg.Colorize {
				colorized = c.cfg.Colorizer.Colorize(label, log.Level)
			}

			dst = append(dst, '[')
			dst = append(dst, colorized...)
			dst = append(dst, ']', ' ')
		}
	}

	if log.Name != "" {
		dst = append(dst, '[')
		dst = append(dst, sanitizeConsoleText(log.Name)...)
		dst = append(dst, ']', ' ')
	}

	dst = append(dst, sanitizeConsoleText(strings.TrimSuffix(log.Message, "\n"))...)

	if log.Caller != "" {
		if len(dst) > 0 && dst[len(dst)-1] != ' ' {
			dst = append(dst, ' ')
		}

		dst = append(dst, CallerKey...)
		dst = append(dst, '=')
		dst = append(dst, sanitizeConsoleText(log.Caller)...)
	}

	var errorText string

	for i := range fields {
		field := fieldAt(i)

		if field.key != ErrorKey || field.isNilAny() {
			continue
		}

		errorText = sanitizeConsoleText(string(field.AppendText(nil)))
	}

	for i := range fields {
		field := fieldAt(i)

		if field.key == "" || field.key == LabelKey || field.key == ErrorKey || field.isNilAny() {
			continue
		}

		if field.key == CallerKey && log.Caller != "" {
			continue
		}

		dst = append(dst, ' ')
		dst = append(dst, sanitizeConsoleText(field.key)...)
		dst = append(dst, '=')
		dst = append(dst, sanitizeConsoleText(string(field.AppendText(nil)))...)
	}

	if errorText != "" {
		dst = append(dst, '\n', '\n')
		dst = append(dst, errorText...)
	}

	return dst, nil
}

// sanitizeConsoleText removes control characters except tab. It prevents
// untrusted text from injecting terminal escapes or extra lines.
//
// Parameters:
//   - s (string): text to sanitize.
//
// Returns:
//   - sanitized (string): text without control characters.
func sanitizeConsoleText(s string) (sanitized string) {
	hasControl := strings.ContainsFunc(s, func(r rune) bool {
		return r != '\t' && unicode.IsControl(r)
	})
	if !hasControl {
		return s
	}

	var builder strings.Builder

	builder.Grow(len(s))

	for _, r := range s {
		if r == '\t' || !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

var defaultLabels = [...]string{
	hqgologgerlevel.LevelFatal: "FTL",
	hqgologgerlevel.LevelError: "ERR",
	hqgologgerlevel.LevelWarn:  "WRN",
	hqgologgerlevel.LevelInfo:  "INF",
	hqgologgerlevel.LevelDebug: "DBG",
	hqgologgerlevel.LevelTrace: "TRC",
}

var (
	_ Formatter = (*Console)(nil)
	_ Appender  = (*Console)(nil)
)
