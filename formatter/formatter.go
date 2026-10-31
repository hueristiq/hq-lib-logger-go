package formatter

import (
	"time"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

// Log is a single log record passed to a [Formatter].
//
// Formatters must not mutate the record or its Fields slice.
//
// Fields:
//   - Time (time.Time): record timestamp; zero means omit it.
//   - Level (hqgologgerlevel.Level): record severity.
//   - Message (string): primary human-readable text.
//   - Caller (string): optional "file.go:line" location.
//   - Name (string): optional dotted logger name, such as "api.v1".
//   - Fields ([]Field): typed context attached to the record.
type Log struct {
	Time    time.Time
	Level   hqgologgerlevel.Level
	Message string
	Caller  string
	Name    string
	Fields  []Field
}

// Formatter converts a Log into bytes.
//
// Implementations must not mutate the record and must be safe for concurrent
// use. The returned slice must remain valid after Format returns.
//
// Methods:
//   - Format(log Log) (data []byte, err error): returns the formatted record.
type Formatter interface {
	Format(log Log) (data []byte, err error)
}

// Appender is an optional Formatter extension that appends to dst. The logger
// uses it to reuse a pooled buffer.
//
// Methods:
//   - Append(dst []byte, log Log) (data []byte, err error): appends the record.
type Appender interface {
	Append(dst []byte, log Log) (data []byte, err error)
}

const (
	// LabelKey is the reserved field key for a short label such as "INF".
	LabelKey = "label"
	// ErrorKey is the reserved field key for an attached error.
	ErrorKey = "error"
	// CallerKey is the reserved field key for the caller annotation.
	CallerKey = "caller"
)
