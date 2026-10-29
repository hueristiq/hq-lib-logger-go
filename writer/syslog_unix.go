//go:build !windows && !plan9

package writer

import (
	"errors"
	"log/syslog"
)

// syslogDial opens a syslog connection.
//
// Parameters:
//   - network (string): "", "unix", "unixgram", "udp", or "tcp".
//   - raddr (string): remote address for udp/tcp.
//   - priority (syslog.Priority): facility.
//   - tag (string): program tag.
//
// Returns:
//   - writer (*syslog.Writer): the connection.
//   - err (error): a dial error.
type syslogDial func(network, raddr string, priority syslog.Priority, tag string) (*syslog.Writer, error)

// NewSyslog connects to the local or a remote syslog daemon. A nil config uses
// the local socket, the tag "hq-lib-logger-go", and LOG_USER.
//
// Parameters:
//   - cfg (*SyslogConfig): connection settings; nil uses defaults.
//
// Returns:
//   - writer (*Syslog): the syslog writer.
//   - err (error): a dial error.
func NewSyslog(cfg *SyslogConfig) (writer *Syslog, err error) {
	return newSyslog(cfg, syslog.Dial)
}

// newSyslog connects to syslog through the given dial function, applying the
// config defaults for tag and facility. The dial indirection keeps the
// constructor testable.
//
// Parameters:
//   - cfg (*SyslogConfig): connection settings; nil uses defaults.
//   - dial (syslogDial): dial function to use.
//
// Returns:
//   - writer (*Syslog): the syslog writer.
//   - err (error): a dial error, or an error for a nil connection.
func newSyslog(cfg *SyslogConfig, dial syslogDial) (writer *Syslog, err error) {
	config := SyslogConfig{}

	if cfg != nil {
		config = *cfg
	}

	if config.Tag == "" {
		config.Tag = "hq-lib-logger-go"
	}

	if config.Facility == 0 {
		config.Facility = syslogFacilityUser
	}

	connection, err := dial(config.Network, config.Address, syslog.Priority(config.Facility), config.Tag)
	if err != nil {
		return nil, err
	}

	if connection == nil {
		return nil, errors.New("writer: syslog dial returned nil connection")
	}

	writer = &Syslog{
		backend:  &syslogBackendAdapter{writer: connection},
		facility: config.Facility,
	}

	return writer, nil
}

// syslogLevelWriter is the severity-method surface of a syslog connection.
//
// Methods:
//   - Crit(message string) (err error): writes a critical message.
//   - Err(message string) (err error): writes an error message.
//   - Warning(message string) (err error): writes a warning message.
//   - Notice(message string) (err error): writes a notice message.
//   - Info(message string) (err error): writes an informational message.
//   - Debug(message string) (err error): writes a debug message.
//   - Close() (err error): releases the connection.
type syslogLevelWriter interface {
	Crit(message string) error
	Err(message string) error
	Warning(message string) error
	Notice(message string) error
	Info(message string) error
	Debug(message string) error
	Close() error
}

// syslogBackendAdapter adapts a syslog connection's severity methods to the
// priority-based syslogBackend interface.
//
// Fields:
//   - writer (syslogLevelWriter): the adapted connection.
type syslogBackendAdapter struct {
	writer syslogLevelWriter
}

// WriteLevel writes a message at the given severity through the adapted
// syslog connection.
//
// Parameters:
//   - priority (int): syslog priority, including severity bits.
//   - message (string): message to write.
//
// Returns:
//   - err (error): a backend error, or [ErrUnsupported] for unknown severity.
func (a *syslogBackendAdapter) WriteLevel(priority int, message string) error {
	switch priority & 0x07 {
	case syslogSeverityCritical:
		return a.writer.Crit(message)
	case syslogSeverityError:
		return a.writer.Err(message)
	case syslogSeverityWarning:
		return a.writer.Warning(message)
	case syslogSeverityNotice:
		return a.writer.Notice(message)
	case syslogSeverityInfo:
		return a.writer.Info(message)
	case syslogSeverityDebug:
		return a.writer.Debug(message)
	default:
		return ErrUnsupported
	}
}

// Close closes the adapted syslog connection.
//
// Returns:
//   - err (error): the connection's close error.
func (a *syslogBackendAdapter) Close() error {
	return a.writer.Close()
}
