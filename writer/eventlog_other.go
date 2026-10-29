//go:build !windows

package writer

// NewEventLog reports that the Windows event log is unavailable on this
// platform.
//
// Parameters:
//   - cfg (*EventLogConfig): ignored.
//
// Returns:
//   - writer (*EventLog): always nil.
//   - err (error): always [ErrUnsupported].
func NewEventLog(_ *EventLogConfig) (writer *EventLog, err error) {
	return nil, ErrUnsupported
}
