//go:build !windows && !plan9

package writer

import (
	"errors"
	"log/syslog"
)

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

type syslogLevelWriter interface {
	Crit(message string) error
	Err(message string) error
	Warning(message string) error
	Notice(message string) error
	Info(message string) error
	Debug(message string) error
	Close() error
}

type syslogBackendAdapter struct {
	writer syslogLevelWriter
}

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

func (a *syslogBackendAdapter) Close() error {
	return a.writer.Close()
}
