package writer

import (
	"io"
	"os"
	"sync"

	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

// Console is a thread-safe implementation of the Writer interface that writes log
// messages to standard output (stdout) or standard error (stderr) based on the log
// level and configuration settings. It supports configurable output destinations
// and newline behavior, making it suitable for console-based logging in various
// environments. The writer uses a mutex to ensure thread-safe access to output
// streams, preventing concurrent write conflicts.
//
// Fields:
//   - mutex (*sync.Mutex): Ensures thread-safe access to stdout and stderr during
//     write operations, preventing data corruption in concurrent environments.
//   - stdout (io.Writer): The output stream for messages directed to standard output,
//     typically os.Stdout but customizable for testing or alternative destinations.
//   - stderr (io.Writer): The output stream for messages directed to standard error,
//     typically os.Stderr but customizable for testing or alternative destinations.
//   - cfg (*ConsoleWriterConfiguration): Configuration settings controlling output
//     destination (stdout/stderr) and newline behavior.
type Console struct {
	mutex  *sync.Mutex
	stdout io.Writer
	stderr io.Writer
	cfg    *ConsoleWriterConfiguration
}

// Write writes the provided log data to either stdout or stderr based on the
// specified log level and configuration settings, appending a newline character
// unless disabled. By default, messages with LevelSilent are written to stdout,
// while all other levels (LevelFatal, LevelError, LevelInfo, LevelWarn, LevelDebug)
// are written to stderr. Configuration options (ForceStderr or ForceStdout) can
// override this behavior to direct all messages to a single stream. The method is
// thread-safe, using a mutex to serialize write operations. If the output stream
// supports flushing (e.g., via a Flush method), it is called to ensure immediate
// output delivery.
//
// Parameters:
//   - data ([]byte): The pre-formatted log message to write, typically produced by
//     a formatter (e.g., as JSON or plain text).
//   - level (hqgologgerlevels.Level): The severity level of the log message, as defined in
//     the levels package (e.g., LevelSilent, LevelError), used to determine the
//     output destination unless overridden by configuration.
//
// Returns:
//   - err (error): An error if writing to the output stream fails (e.g., due to I/O
//     issues) or if flushing fails for a flushable stream. Returns nil if the write
//     and optional flush operations succeed.
func (c *Console) Write(data []byte, level hqgologgerlevels.Level) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var writer io.Writer

	switch {
	case c.cfg.ForceStderr:
		writer = c.stderr
	case c.cfg.ForceStdout:
		writer = c.stdout
	case level == hqgologgerlevels.LevelSilent:
		writer = c.stdout
	default:
		writer = c.stderr
	}

	if _, err = writer.Write(data); err != nil {
		return
	}

	if !c.cfg.DisableNewline {
		if _, err = writer.Write([]byte("\n")); err != nil {
			return
		}
	}

	if flusher, ok := writer.(interface{ Flush() error }); ok {
		if err = flusher.Flush(); err != nil {
			return
		}

		return
	}

	return
}

// Close closes the stdout and stderr streams if they are not os.Stdout or os.Stderr
// and implement the io.Closer interface. This ensures proper resource cleanup for
// custom output streams (e.g., file handles or network connections used in testing).
// The method is thread-safe, using a mutex to prevent concurrent access. It attempts
// to close both streams and returns the last non-nil error encountered, if any.
// If the streams are os.Stdout or os.Stderr, they are not closed, as these are
// managed by the operating system.
//
// Returns:
//   - err (error): The last non-nil error from closing either stream, or nil if
//     both streams are closed successfully or are not closable (e.g., os.Stdout).
func (c *Console) Close() (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.stdout != os.Stdout {
		if closer, ok := c.stdout.(io.Closer); ok {
			err = closer.Close()
		}
	}

	if c.stderr != os.Stderr {
		if closer, ok := c.stderr.(io.Closer); ok {
			err = closer.Close()
		}
	}

	return
}

// ConsoleWriterConfiguration defines configuration options for the Console writer.
// It allows customization of output destination and newline behavior to adapt the
// writer to different logging requirements.
//
// Fields:
//   - ForceStderr (bool): If true, directs all log messages to stderr, overriding
//     the default behavior of routing LevelSilent to stdout.
//   - ForceStdout (bool): If true, directs all log messages to stdout, overriding
//     the default behavior of routing non-silent levels to stderr.
//   - DisableNewline (bool): If true, prevents appending a newline character to
//     each log message, useful for custom formatting or when newlines are handled
//     by the formatter.
type ConsoleWriterConfiguration struct {
	ForceStderr    bool
	ForceStdout    bool
	DisableNewline bool
}

var _ Writer = (*Console)(nil)

// DefaultConsoleWriterConfig returns a default configuration for the Console writer.
// The default settings direct LevelSilent messages to stdout, other levels to stderr,
// and append a newline to each message. This provides a sensible starting point for
// console logging that can be customized as needed.
//
// Returns:
//   - cfg (*ConsoleWriterConfiguration): A pointer to the default configuration.
func DefaultConsoleWriterConfig() (cfg *ConsoleWriterConfiguration) {
	cfg = &ConsoleWriterConfiguration{
		ForceStderr:    false,
		ForceStdout:    false,
		DisableNewline: false,
	}

	return
}

// NewConsoleWriter creates and returns a new Console writer instance, initialized
// with a mutex for thread-safe operation and the provided configuration. If no
// configuration is provided (i.e., cfg is nil), it uses the default configuration
// from DefaultConsoleWriterConfig. The writer uses os.Stdout and os.Stderr as
// default output streams but allows customization for testing or alternative
// destinations. The instance is ready for use in a logging system to write
// formatted log messages to console outputs.
//
// Parameters:
//   - cfg (*ConsoleWriterConfiguration): The configuration for the writer. If nil,
//     defaults are applied.
//
// Returns:
//   - writer (*Console): A pointer to a new Console writer instance.
func NewConsoleWriter(cfg *ConsoleWriterConfiguration) (writer *Console) {
	if cfg == nil {
		cfg = DefaultConsoleWriterConfig()
	}

	writer = &Console{
		mutex:  &sync.Mutex{},
		stdout: os.Stdout,
		stderr: os.Stderr,
		cfg:    cfg,
	}

	return
}
