package writer

import (
	"io"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// MultiWriter is an implementation of the Writer interface that aggregates
// multiple Writer instances, forwarding log messages to each underlying writer.
// It enables simultaneous logging to multiple destinations (e.g., console and file)
// while maintaining a single Writer interface. Errors from individual writers are
// collected, and the last non-nil error is returned, ensuring all writers are
// attempted even if some fail. Nil writers are filtered out during initialization
// to prevent runtime issues.
//
// Fields:
//   - writers ([]Writer): The slice of Writer instances to which log messages are
//     forwarded. Each writer handles its own output destination and level filtering.
type MultiWriter struct {
	writers []Writer
}

// Write forwards the provided log data and severity level to each underlying
// Writer in the MultiWriter's writers slice. It attempts to write to all writers,
// even if some fail, and returns the last non-nil error encountered (if any).
// This ensures that partial failures do not prevent other writers from processing
// the log message. The method is thread-safe as long as the underlying writers
// are thread-safe.
//
// Parameters:
//   - data ([]byte): The pre-formatted log message to write, typically produced
//     by a formatter.
//   - level (hqgologgerlevels.Level): The severity level of the log message, as defined in
//     the levels package (e.g., LevelFatal, LevelDebug), used by writers to filter
//     messages based on their configured thresholds.
//
// Returns:
//   - err (error): The last non-nil error from any underlying writer, or nil if
//     all writes succeed or no writers are present.
func (m *MultiWriter) Write(data []byte, level hqgologgerlevels.Level) (err error) {
	for _, writer := range m.writers {
		err = writer.Write(data, level)
	}

	return
}

// Close closes all underlying writers in the MultiWriter's writers slice,
// releasing their associated resources. It attempts to close all writers, even
// if some fail, and returns the last non-nil error encountered (if any). This
// ensures that partial failures do not prevent other writers from closing
// properly. The method is thread-safe as long as the underlying writers are
// thread-safe.
//
// Returns:
//   - err (error): The last non-nil error from any underlying writer, or nil if
//     all closes succeed or no writers are present.
func (m *MultiWriter) Close() (err error) {
	for _, writer := range m.writers {
		err = writer.Close()
	}

	return
}

// Writer defines the interface for writing log messages to an output destination.
// Implementations of this interface handle the delivery of formatted log data to
// specific sinks, such as files, consoles, network endpoints, or external logging
// services, while considering the severity level of the log message. The interface
// extends io.Closer to ensure resources (e.g., file handles or network connections)
// can be properly closed when logging is complete.
//
// Methods:
//   - Write(data []byte, level hqgologgerlevels.Level) (err error): Writes the provided log
//     data to the output destination if the severity level meets the writer's
//     criteria (e.g., a configured threshold). The data is typically pre-formatted
//     by a formatter (e.g., as JSON or plain text), and the level is a severity
//     value from the levels package (e.g., LevelInfo, LevelError). Returns an error
//     if the write operation fails (e.g., due to I/O issues or destination
//     unavailability).
//   - Close() (err error): Closes the writer, releasing any associated resources
//     (e.g., file handles or network connections). Returns an error if the close
//     operation fails.
type Writer interface {
	io.Closer
	Write(data []byte, level hqgologgerlevels.Level) (err error)
}

var _ Writer = (*MultiWriter)(nil)

// NewMultiWriter creates and returns a new MultiWriter instance that aggregates
// the provided Writer instances. It filters out nil writers to ensure safe
// operation and initializes the writers slice with the non-nil writers. The
// resulting MultiWriter can be used to forward log messages to multiple
// destinations simultaneously. If no non-nil writers are provided, an empty
// MultiWriter is returned, which performs no operations when used.
//
// Parameters:
//   - writers (...Writer): A variadic list of Writer instances to aggregate.
//
// Returns:
//   - multi (*MultiWriter): A pointer to a new MultiWriter instance containing
//     the non-nil writers.
func NewMultiWriter(writers ...Writer) (multi *MultiWriter) {
	multi = &MultiWriter{
		writers: make([]Writer, 0, len(writers)),
	}

	for _, writter := range writers {
		if writter != nil {
			multi.writers = append(multi.writers, writter)
		}
	}

	return
}
