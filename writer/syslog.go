package writer

import (
	"fmt"
	"strings"
	"sync"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

const (
	syslogFacilityUser = 1 << 3

	syslogSeverityCritical = 2
	syslogSeverityError    = 3
	syslogSeverityWarning  = 4
	syslogSeverityNotice   = 5
	syslogSeverityInfo     = 6
	syslogSeverityDebug    = 7
)

// SyslogConfig configures a Unix syslog writer.
//
// Fields:
//   - Network (string): "", "unix", "unixgram", "udp", or "tcp".
//   - Address (string): remote address for udp/tcp.
//   - Tag (string): program tag; empty means "hq-lib-logger-go".
//   - Facility (int): syslog facility; zero means LOG_USER.
type SyslogConfig struct {
	Network  string
	Address  string
	Tag      string
	Facility int
}

// syslogBackend abstracts the syslog connection, keeping the writer
// testable.
//
// Methods:
//   - WriteLevel(priority int, message string) (err error): writes a message
//     at the given priority.
//   - Close() (err error): releases the connection.
type syslogBackend interface {
	WriteLevel(priority int, message string) error
	Close() error
}

// Syslog writes records to Unix syslog. The constructor returns
// [ErrUnsupported] on platforms without syslog.
//
// Fields:
//   - mutex (sync.Mutex): serializes writes and Close.
//   - backend (syslogBackend): platform syslog connection.
//   - facility (int): configured syslog facility.
//   - closed (bool): whether Close has run.
//   - closeErr (error): result of the first Close.
type Syslog struct {
	mutex    sync.Mutex
	backend  syslogBackend
	facility int
	closed   bool
	closeErr error
}

// Write sends data at the severity mapped from level.
//
// Parameters:
//   - data ([]byte): formatted record.
//   - level (hqgologgerlevel.Level): record severity.
//
// Returns:
//   - err (error): [ErrUnsupported], [ErrClosed], or a backend error.
func (s *Syslog) Write(data []byte, level hqgologgerlevel.Level) (err error) {
	if s == nil || s.backend == nil {
		return ErrUnsupported
	}

	if !level.Valid() {
		return fmt.Errorf("%w (%d)", hqgologgerlevel.ErrUnknownLevel, int(level))
	}

	severity, ok := syslogSeverity(level)
	if !ok {
		// Unreachable: the Valid check above admits only levels the
		// mapping covers. Kept defensively.
		return fmt.Errorf("%w (%d)", hqgologgerlevel.ErrUnknownLevel, int(level))
	}

	facility := s.facility
	if facility == 0 {
		facility = syslogFacilityUser
	}

	message := strings.TrimRight(string(data), "\r\n")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return ErrClosed
	}

	return s.backend.WriteLevel(facility|severity, message)
}

// Close closes the syslog connection. Repeated calls return the first error.
//
// Returns:
//   - err (error): the backend's close error, if any.
func (s *Syslog) Close() (err error) {
	if s == nil || s.backend == nil {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return s.closeErr
	}

	s.closed = true
	s.closeErr = s.backend.Close()

	return s.closeErr
}

// syslogSeverity maps a log level to the syslog severity bits 0-7.
//
// Parameters:
//   - level (hqgologgerlevel.Level): severity to map.
//
// Returns:
//   - severity (int): the syslog severity.
//   - ok (bool): true when the level is known.
func syslogSeverity(level hqgologgerlevel.Level) (severity int, ok bool) {
	switch level {
	case hqgologgerlevel.LevelFatal:
		return syslogSeverityCritical, true
	case hqgologgerlevel.LevelError:
		return syslogSeverityError, true
	case hqgologgerlevel.LevelWarn:
		return syslogSeverityWarning, true
	case hqgologgerlevel.LevelInfo:
		return syslogSeverityInfo, true
	case hqgologgerlevel.LevelDebug, hqgologgerlevel.LevelTrace:
		return syslogSeverityDebug, true
	default:
		return 0, false
	}
}

var _ Writer = (*Syslog)(nil)
