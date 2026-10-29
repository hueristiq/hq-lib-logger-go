//go:build windows || plan9

package writer

// NewSyslog reports that syslog is unavailable on this platform.
//
// Parameters:
//   - cfg (*SyslogConfig): ignored.
//
// Returns:
//   - writer (*Syslog): always nil.
//   - err (error): always [ErrUnsupported].
func NewSyslog(_ *SyslogConfig) (writer *Syslog, err error) {
	return nil, ErrUnsupported
}
