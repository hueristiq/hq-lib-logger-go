package formatter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Console is an implementation of the Formatter interface that formats log messages
// for console output. It constructs a string in the format "[timestamp] [label] message [metadata]"
// (with optional components based on configuration) and returns it as a byte slice.
// Timestamps, labels, and metadata are included based on the configuration settings.
// Labels are colorized using the provided Colorizer if enabled. Metadata is appended
// as key=value pairs, with special handling for errors to include stack traces when
// applicable. The output is optimized for human-readable console display and does
// not include a trailing newline, as this is typically handled by the log writer.
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
// Labels are extracted from metadata and colorized if enabled. The message is trimmed
// of trailing newlines. Metadata is appended as key=value pairs, with the "error" entry
// rendered as a trailing block containing the error message. The buffer is pre-allocated
// with an estimated size for efficiency.
//
// Parameters:
//   - log (*Log): The log message to format, containing context, timestamp, level,
//     message, and optional metadata.
//
// Returns:
//   - data ([]byte): The formatted log message as a byte slice, ready for console output.
//   - err (error): An error if the log level is invalid, otherwise nil.
func (c *Console) Format(log *Log) (data []byte, err error) {
	if !log.Level.IsValid() {
		err = fmt.Errorf("invalid log level: %d", log.Level.Int())

		return
	}

	metadata := make(map[string]interface{})

	for k, v := range log.Metadata {
		metadata[k] = v
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

	if label, ok := metadata["label"]; ok {
		if label != "" && c.cfg.IncludeLabel {
			if str, ok := label.(string); ok && str != "" {
				colorized := str

				if c.cfg.Colorize {
					colorized = c.cfg.Colorizer.Colorize(str, log.Level)
				}

				buffer.WriteByte('[')
				buffer.WriteString(colorized)
				buffer.WriteByte(']')
				buffer.WriteByte(' ')
			}
		}

		delete(metadata, "label")
	}

	message := strings.TrimSuffix(log.Message, "\n")

	buffer.WriteString(message)

	var formattedErrorMetadata string

	if errValue, ok := metadata["error"]; ok && errValue != nil {
		if err, ok := errValue.(error); ok {
			formattedErrorMetadata = "\n\n" + err.Error()
		} else {
			formattedErrorMetadata = fmt.Sprintf("\n\n%v", errValue)
		}

		delete(metadata, "error")
	}

	keys := make([]string, 0, len(metadata))

	for k := range metadata {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := metadata[k]

		if k == "" || v == nil {
			continue
		}

		buffer.WriteByte(' ')
		buffer.WriteString(k)
		buffer.WriteByte('=')

		fmt.Fprintf(buffer, "%v", v)
	}

	buffer.WriteString(formattedErrorMetadata)

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
//   - IncludeLabel (bool): If true, includes a label (from metadata["label"]) in the output.
//   - Colorize (bool): If true, enables colorization of labels using the Colorizer.
//   - Colorizer (Colorizer): The Colorizer implementation used for applying colors to labels.
//   - PrettyPrint (bool): If true, enables pretty-printing of output (currently unused).
type ConsoleFormatterConfiguration struct {
	IncludeTimestamp bool
	TimestampFormat  string
	IncludeLabel     bool
	Colorize         bool
	Colorizer        Colorizer
	PrettyPrint      bool
}

var _ Formatter = (*Console)(nil)

// DefaultConsoleConfig returns a default configuration for the Console formatter.
// The default settings include a timestamp in RFC3339 format, label inclusion,
// colorization with a no-op Colorizer, and disable pretty-printing. This provides
// a sensible starting point for console logging that can be customized as needed.
//
// Returns:
//   - cfg (*ConsoleFormatterConfiguration): A pointer to the default configuration.
func DefaultConsoleConfig() (cfg *ConsoleFormatterConfiguration) {
	cfg = &ConsoleFormatterConfiguration{
		IncludeTimestamp: true,
		TimestampFormat:  time.RFC3339,
		IncludeLabel:     true,
		Colorize:         true,
		Colorizer:        NewNoOpColorizer(),
		PrettyPrint:      false,
	}

	return
}

// NewConsoleFormatter creates and returns a new Console formatter instance,
// configured with the provided ConsoleFormatterConfiguration. If no configuration
// is provided (i.e., cfg is nil), it uses the default configuration from
// DefaultConsoleConfig. This factory function ensures the formatter is properly
// initialized for use in logging systems.
//
// Parameters:
//   - cfg (*ConsoleFormatterConfiguration): The configuration for the formatter.
//     If nil, defaults are applied.
//
// Returns:
//   - formatter (*Console): A pointer to a new Console formatter instance.
func NewConsoleFormatter(cfg *ConsoleFormatterConfiguration) (formatter *Console) {
	if cfg == nil {
		cfg = DefaultConsoleConfig()
	}

	formatter = &Console{
		cfg: cfg,
	}

	return
}
