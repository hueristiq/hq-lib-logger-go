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
// streams, preventing concurrent write conflicts. Construct with
// [NewConsoleWriter]; see [ConsoleWriterConfiguration] for the available options.
// A Console must not be copied after first use.
//
// Fields:
//   - mutex (sync.Mutex): Ensures thread-safe access to stdout and stderr during
//     write operations, preventing data corruption in concurrent environments.
//   - stdout (io.Writer): The output stream for messages directed to standard output,
//     typically os.Stdout but customizable via ConsoleWriterConfiguration.Stdout for
//     testing or alternative destinations.
//   - stderr (io.Writer): The output stream for messages directed to standard error,
//     typically os.Stderr but customizable via ConsoleWriterConfiguration.Stderr for
//     testing or alternative destinations.
//   - cfg (*ConsoleWriterConfiguration): Configuration settings controlling output
//     destination (stdout/stderr) and newline behavior.
//   - buf ([]byte): Scratch space reused across writes to append the trailing
//     newline without a per-line allocation. Guarded by mutex; it grows to, and
//     retains, the size of the largest line written.
type Console struct {
	mutex  sync.Mutex
	stdout io.Writer
	stderr io.Writer
	cfg    *ConsoleWriterConfiguration
	buf    []byte
}

// Write writes the provided log data to either stdout or stderr based on the
// specified log level and configuration settings, appending a newline character
// unless disabled. By default, messages with LevelSilent are written to stdout,
// while all other levels (LevelFatal, LevelError, LevelInfo, LevelWarn, LevelDebug)
// are written to stderr. Configuration options (ForceStderr or ForceStdout) can
// override this behavior to direct all messages to a single stream. The method is
// thread-safe, using a mutex to serialize write operations. If the output stream
// supports flushing (e.g., via a Flush method), it is called to ensure immediate
// output delivery. The newline is appended in a reusable scratch buffer, so the
// caller's slice — including its spare capacity — is never modified, and the payload
// and its newline are still delivered in a single write. The scratch buffer is
// retained between writes and grows to fit the largest line seen, trading a bounded
// memory footprint for zero per-write allocations.
//
// Parameters:
//   - data ([]byte): The pre-formatted log message to write, typically produced by
//     a formatter.
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

	var w io.Writer

	switch {
	case c.cfg.ForceStderr:
		w = c.stderr
	case c.cfg.ForceStdout:
		w = c.stdout
	case level == hqgologgerlevels.LevelSilent:
		w = c.stdout
	default:
		w = c.stderr
	}

	if !c.cfg.DisableNewline {
		// Reuse the scratch buffer instead of allocating a per-line copy. Safe
		// because Write is fully serialized by the mutex and the io.Writer
		// contract forbids the destination from retaining the slice it is given.
		c.buf = append(c.buf[:0], data...)
		c.buf = append(c.buf, '\n')

		data = c.buf
	}

	if _, err = w.Write(data); err != nil {
		return
	}

	if flusher, ok := w.(interface{ Flush() error }); ok {
		err = flusher.Flush()
	}

	return
}

// Close closes the stdout and stderr streams if they are not os.Stdout or os.Stderr
// and implement the io.Closer interface. This ensures proper resource cleanup for
// custom output streams (e.g., file handles or network connections injected via
// ConsoleWriterConfiguration). The method is thread-safe, using a mutex to prevent
// concurrent access. Both streams are attempted even if one fails, and the last
// non-nil error is returned. If the streams are os.Stdout or os.Stderr, they are
// not closed, as these are managed by the operating system.
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
			if cerr := closer.Close(); cerr != nil {
				err = cerr
			}
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
//   - Stdout (io.Writer): The stream for messages routed to standard output. If nil,
//     os.Stdout is used. Custom streams are closed by Close if they implement
//     io.Closer.
//   - Stderr (io.Writer): The stream for messages routed to standard error. If nil,
//     os.Stderr is used. Custom streams are closed by Close if they implement
//     io.Closer.
type ConsoleWriterConfiguration struct {
	ForceStderr    bool
	ForceStdout    bool
	DisableNewline bool
	Stdout         io.Writer
	Stderr         io.Writer
}

// Compile-time guard ensuring [Console] satisfies [Writer].
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
// for thread-safe operation with the provided configuration. If no configuration
// is provided (i.e., cfg is nil), it uses the default configuration from
// DefaultConsoleWriterConfig. The writer uses os.Stdout and os.Stderr as default
// output streams; cfg.Stdout and cfg.Stderr can override them for testing or
// alternative destinations. The configuration is copied before use, so mutating
// the caller's struct afterwards does not affect the writer. The instance is
// ready for use in a logging system to write formatted log messages to console
// outputs.
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

	copied := *cfg

	writer = &Console{
		stdout: os.Stdout,
		stderr: os.Stderr,
		cfg:    &copied,
	}

	if copied.Stdout != nil {
		writer.stdout = copied.Stdout
	}

	if copied.Stderr != nil {
		writer.stderr = copied.Stderr
	}

	return
}
