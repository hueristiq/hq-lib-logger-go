// Package formatter renders log records into bytes for output.
//
// # Entry points
//
// A [Formatter] converts a [Log] into a byte slice. [NewConsole] produces
// human-readable lines, [NewJSON] emits one JSON object per record, and
// [NewLogfmt] emits compact key=value lines. All three consume opaque [Field]
// values and implement [Appender].
//
// # Configuration
//
// Each constructor accepts a nil config for package defaults. A non-nil zero
// config is explicit: optional keys and timestamps are off unless enabled.
//
// # Concurrency
//
// Formatters are safe for concurrent use and never mutate the input record.
// Formatters receive the record by value and must not mutate it.
package formatter
