package formatter

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// Console is an implementation of the Formatter interface that formats log messages
// for console output. It constructs a string in the format "[timestamp] [label] message [metadata]"
// (with optional components based on configuration) and returns it as a byte slice.
// Timestamps, labels, and metadata are included based on the configuration settings.
// When a log carries no label of its own, a default label based on the level is
// applied (e.g., "INF" for LevelInfo); an explicitly empty label suppresses the
// tag entirely. Labels are colorized using the provided Colorizer if enabled.
// Metadata is appended as key=value pairs in sorted key order for stable output,
// and the "error" entry is rendered as a trailing block containing the error's
// message. The output is optimized for human-readable console display and does
// not include a trailing newline, as this is typically handled by the log writer.
// Construct with [NewConsoleFormatter]; see [ConsoleFormatterConfiguration] for
// the available options.
//
// Fields:
//   - cfg (*ConsoleFormatterConfiguration): Configuration settings for the formatter,
//     controlling timestamp inclusion, label usage, colorization, and metadata handling.
type Console struct {
	cfg *ConsoleFormatterConfiguration
}

// Format converts a Log struct into a formatted byte slice for console output.
// The output format is "[timestamp] [label] message [metadata]" (with optional components).
// Timestamps are included if configured, using the specified format (default: RFC3339).
// The label is taken from metadata[LabelKey] and colorized if enabled; when the
// key is absent, the level's default label (see defaultLabels) is used instead,
// while an explicitly empty or non-string label omits the tag. The message is
// trimmed of a single trailing newline. Metadata is appended as key=value pairs,
// with the reserved keys [LabelKey] and [ErrorKey] skipped — the ErrorKey entry
// is instead rendered as a trailing block containing the error message, separated
// from the message by a blank line. The input Log and its Metadata map are never
// mutated. The buffer is pre-allocated with an estimated size for efficiency.
// The returned slice is freshly allocated for each call and stays valid after
// Format returns, as the [Formatter] contract requires.
//
// Parameters:
//   - log (*Log): The log message to format, containing timestamp, level, message,
//     and optional metadata.
//
// Returns:
//   - data ([]byte): The formatted log message as a byte slice, ready for console output.
//   - err (error): An error wrapping
//     [github.com/hueristiq/hq-lib-logger-go/levels.ErrUnknownLevel] if the log
//     level is invalid; otherwise nil.
func (c *Console) Format(log *Log) (data []byte, err error) {
	if !log.Level.IsValid() {
		err = fmt.Errorf("%w (%d)", hqgologgerlevels.ErrUnknownLevel, log.Level.Int())

		return
	}

	buffer := &bytes.Buffer{}

	estimatedSize := len(log.Message) + 50

	if c.cfg.IncludeTimestamp {
		estimatedSize += 25
	}

	if c.cfg.IncludeLabel {
		estimatedSize += 10
	}

	buffer.Grow(estimatedSize)

	if c.cfg.IncludeTimestamp && !log.Timestamp.IsZero() {
		buffer.WriteString(log.Timestamp.Format(c.cfg.TimestampFormat))
		buffer.WriteByte(' ')
	}

	if c.cfg.IncludeLabel {
		var label string

		if value, exists := log.Metadata[LabelKey]; exists {
			label, _ = value.(string)
		} else {
			label = defaultLabels[log.Level]
		}

		if label != "" {
			colorized := label

			if c.cfg.Colorize {
				colorized = c.cfg.Colorizer.Colorize(label, log.Level)
			}

			buffer.WriteByte('[')
			buffer.WriteString(colorized)
			buffer.WriteByte(']')
			buffer.WriteByte(' ')
		}
	}

	buffer.WriteString(strings.TrimSuffix(log.Message, "\n"))

	var errorText string

	if errValue, ok := log.Metadata[ErrorKey]; ok && errValue != nil {
		switch e := errValue.(type) {
		case string:
			errorText = e
		case error:
			errorText = e.Error()
		default:
			errorText = fmt.Sprintf("%v", e)
		}
	}

	keys := make([]string, 0, len(log.Metadata))

	for k := range log.Metadata {
		if k == "" || k == LabelKey || k == ErrorKey {
			continue
		}

		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := log.Metadata[k]

		if v == nil {
			continue
		}

		buffer.WriteByte(' ')
		buffer.WriteString(k)
		buffer.WriteByte('=')

		// strconv fast paths byte-match the "%v" verb for the common scalar
		// types, avoiding fmt's reflection. The Append* variants format into
		// stack scratch (digits), so no heap string is allocated per value.
		var digits [32]byte

		switch value := v.(type) {
		case string:
			buffer.WriteString(value)
		case error:
			buffer.WriteString(value.Error())
		case int:
			buffer.Write(strconv.AppendInt(digits[:0], int64(value), 10))
		case int8:
			buffer.Write(strconv.AppendInt(digits[:0], int64(value), 10))
		case int16:
			buffer.Write(strconv.AppendInt(digits[:0], int64(value), 10))
		case int32:
			buffer.Write(strconv.AppendInt(digits[:0], int64(value), 10))
		case int64:
			buffer.Write(strconv.AppendInt(digits[:0], value, 10))
		case uint:
			buffer.Write(strconv.AppendUint(digits[:0], uint64(value), 10))
		case uint8:
			buffer.Write(strconv.AppendUint(digits[:0], uint64(value), 10))
		case uint16:
			buffer.Write(strconv.AppendUint(digits[:0], uint64(value), 10))
		case uint32:
			buffer.Write(strconv.AppendUint(digits[:0], uint64(value), 10))
		case uint64:
			buffer.Write(strconv.AppendUint(digits[:0], value, 10))
		case float64:
			buffer.Write(strconv.AppendFloat(digits[:0], value, 'g', -1, 64))
		case bool:
			buffer.WriteString(strconv.FormatBool(value))
		case time.Duration:
			buffer.WriteString(value.String())
		default:
			fmt.Fprintf(buffer, "%v", v)
		}
	}

	if errorText != "" {
		buffer.WriteString("\n\n")
		buffer.WriteString(errorText)
	}

	data = buffer.Bytes()

	return
}

// ConsoleFormatterConfiguration defines configuration options for the Console formatter.
// It controls the inclusion and formatting of timestamps, labels, metadata, and
// colorization, allowing customization of the console output format.
//
// Fields:
//   - IncludeTimestamp (bool): If true, includes a timestamp in the formatted output.
//   - TimestampFormat (string): The format for timestamps (e.g., time.RFC3339).
//   - IncludeLabel (bool): If true, includes a label (from metadata[LabelKey], or the
//     level's default label when the key is absent) in the output.
//   - Colorize (bool): If true, enables colorization of labels using the Colorizer.
//   - Colorizer (Colorizer): The Colorizer implementation used for applying colors to
//     labels. If nil while Colorize is true, NewConsoleFormatter substitutes a
//     NoOpColorizer, so a configuration never panics on a nil Colorizer.
type ConsoleFormatterConfiguration struct {
	IncludeTimestamp bool
	TimestampFormat  string
	IncludeLabel     bool
	Colorize         bool
	Colorizer        Colorizer
}

// Compile-time guard ensuring [Console] satisfies [Formatter].
var _ Formatter = (*Console)(nil)

// defaultLabels holds, indexed by Level value, the short label applied when a log
// carries no label of its own (no metadata[LabelKey] entry). LevelSilent has no
// default label, so Print output renders without a bracketed tag unless one is
// set explicitly; indices without an entry yield the zero value (""). The level
// is validated by [Console.Format] before lookup, so indexing is always in bounds.
var defaultLabels = [...]string{
	hqgologgerlevels.LevelFatal: "FTL",
	hqgologgerlevels.LevelError: "ERR",
	hqgologgerlevels.LevelInfo:  "INF",
	hqgologgerlevels.LevelWarn:  "WRN",
	hqgologgerlevels.LevelDebug: "DBG",
}

// DefaultConsoleFormatterConfig returns a default configuration for the Console
// formatter. The default settings include a timestamp in RFC3339 format, label
// inclusion, and colorization with a no-op Colorizer. This provides a sensible
// starting point for console logging that can be customized as needed.
//
// Returns:
//   - cfg (*ConsoleFormatterConfiguration): A pointer to the default configuration.
func DefaultConsoleFormatterConfig() (cfg *ConsoleFormatterConfiguration) {
	cfg = &ConsoleFormatterConfiguration{
		IncludeTimestamp: true,
		TimestampFormat:  time.RFC3339,
		IncludeLabel:     true,
		Colorize:         true,
		Colorizer:        NewNoOpColorizer(),
	}

	return
}

// NewConsoleFormatter creates and returns a new Console formatter instance,
// configured with the provided ConsoleFormatterConfiguration. If no configuration
// is provided (i.e., cfg is nil), it uses the default configuration from
// DefaultConsoleFormatterConfig. If the configuration enables Colorize but
// provides no Colorizer, a NoOpColorizer is substituted so formatting cannot
// panic. The configuration is copied before use, so mutating the caller's struct
// afterwards does not affect the formatter. This factory function ensures the
// formatter is properly initialized for use in logging systems.
//
// Parameters:
//   - cfg (*ConsoleFormatterConfiguration): The configuration for the formatter.
//     If nil, defaults are applied.
//
// Returns:
//   - formatter (*Console): A pointer to a new Console formatter instance.
func NewConsoleFormatter(cfg *ConsoleFormatterConfiguration) (formatter *Console) {
	if cfg == nil {
		cfg = DefaultConsoleFormatterConfig()
	}

	copied := *cfg

	if copied.Colorize && copied.Colorizer == nil {
		copied.Colorizer = NewNoOpColorizer()
	}

	formatter = &Console{
		cfg: &copied,
	}

	return
}
