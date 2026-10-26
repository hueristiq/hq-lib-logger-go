package formatter

import (
	"time"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// Log represents a single log message with its associated severity level, content,
// timestamp, and optional metadata. It serves as the input to Formatter
// implementations to produce formatted output for logging systems. The metadata
// keys [LabelKey] and [ErrorKey] are reserved and rendered specially.
//
// Fields:
//   - Timestamp (time.Time): time at which the log message was created.
//   - Level (hqgologgerlevels.Level): severity of the log message.
//   - Message (string): primary human-readable content.
//   - Metadata (map[string]any): optional context; formatters must not mutate it.
type Log struct {
	Timestamp time.Time
	Level     hqgologgerlevels.Level
	Message   string
	Metadata  map[string]any
}

// Formatter defines the interface for formatting log messages. Implementations
// of this interface convert a Log struct into a byte slice, enabling output in
// various formats such as JSON, plain text, or structured logging formats like
// Logfmt. This abstraction allows logging systems to support multiple output
// styles while maintaining a consistent input structure. Implementations must
// not mutate the input [Log] and are expected to be safe for concurrent use,
// since a Logger may format events from several goroutines at once.
//
// The returned slice must remain valid after Format returns: the Logger writes
// it to the output destination after Format has returned. Implementations must
// therefore not recycle the returned slice — for example by returning a
// sync.Pool-owned buffer's bytes — because recycling it corrupts output that
// is still in flight.
//
// Methods:
//   - Format(log *Log) (data []byte, err error): converts log into the formatted
//     bytes, returning an error when formatting fails.
type Formatter interface {
	// Format returns the formatted bytes for log.
	//
	// Parameters:
	//   - log (*Log): the event to format.
	//
	// Returns:
	//   - data ([]byte): the formatted output.
	//   - err (error): non-nil when formatting fails.
	Format(log *Log) (data []byte, err error)
}

const (
	// LabelKey is the metadata key carrying a log event's short label (e.g.,
	// "INF"), rendered as a bracketed tag by the console formatter.
	LabelKey = "label"
	// ErrorKey is the metadata key carrying an error attached to a log event,
	// rendered as a trailing block by the console formatter.
	ErrorKey = "error"
)
