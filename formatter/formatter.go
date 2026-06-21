// Package formatter renders log events into bytes for output.
//
// A [Formatter] converts a [Log] into a byte slice. [Console] is the bundled
// implementation, producing human-readable "[timestamp] [label] message metadata"
// lines with sorted metadata and an optional trailing error block. Label
// colorization is delegated to a [Colorizer]: the default [NoOpColorizer] leaves
// text unchanged, while the
// [github.com/hueristiq/hq-lib-logger-go/formatter/colorizer] package supplies
// ANSI-coloring implementations. Implement [Formatter] to emit other layouts such
// as JSON or Logfmt.
package formatter

import (
	"time"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// Log represents a single log message with its associated severity level, content,
// timestamp, and optional metadata. It serves as the input to Formatter
// implementations to produce formatted output for logging systems.
//
// Fields:
//   - Timestamp (time.Time): The time at which the log message was created,
//     providing a precise record of when the event occurred. Typically used to
//     include timestamps in formatted log output.
//   - Level (hqgologgerlevels.Level): The severity of the log message, as defined in the
//     levels package (e.g., LevelFatal, LevelInfo, LevelDebug). Lower integer
//     values indicate higher severity, influencing how the message is prioritized
//     or filtered.
//   - Message (string): The primary content of the log message, describing the
//     event, condition, or error being logged. This is the main human-readable
//     part of the log.
//   - Metadata (map[string]interface{}): Optional key-value pairs providing
//     additional context for the log message. Metadata can include structured
//     data such as request IDs, user IDs, system metrics, or other relevant
//     information to aid in debugging or analysis. The use of interface{} allows
//     flexibility in the types of values stored.
type Log struct {
	Timestamp time.Time
	Level     hqgologgerlevels.Level
	Message   string
	Metadata  map[string]interface{}
}

// Formatter defines the interface for formatting log messages. Implementations
// of this interface convert a Log struct into a byte slice, enabling output in
// various formats such as JSON, plain text, or structured logging formats like
// Logfmt. This abstraction allows logging systems to support multiple output
// styles while maintaining a consistent input structure.
//
// Methods:
//   - Format(log *Log) (data []byte, err error): Converts the provided Log
//     struct into a byte slice representing the formatted log message. The
//     formatted output is suitable for writing to an output destination, such as
//     a file, console, or network stream. Returns an error if formatting fails,
//     for example, due to serialization issues in structured formats like JSON.
//     Implementations should handle all fields of the Log struct appropriately,
//     including optional fields like Metadata and Context.
type Formatter interface {
	Format(log *Log) (data []byte, err error)
}
