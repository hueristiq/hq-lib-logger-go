//go:build windows

package writer

import (
	"errors"

	"golang.org/x/sys/windows/svc/eventlog"
)

// eventLogOpen opens a Windows event source by name.
//
// Parameters:
//   - source (string): registered event source name.
//
// Returns:
//   - log (*eventlog.Log): the opened handle.
//   - err (error): an open error.
type eventLogOpen func(source string) (*eventlog.Log, error)

// NewEventLog opens a registered Windows event source. The source must already
// be installed.
//
// Parameters:
//   - cfg (*EventLogConfig): source and event ID; Source is required.
//
// Returns:
//   - writer (*EventLog): the event log writer.
//   - err (error): a configuration or open error.
func NewEventLog(cfg *EventLogConfig) (writer *EventLog, err error) {
	return newEventLog(cfg, eventlog.Open)
}

// newEventLog opens a Windows event source through the given open function.
// The open indirection keeps the constructor testable.
//
// Parameters:
//   - cfg (*EventLogConfig): source and event ID; Source is required.
//   - open (eventLogOpen): open function to use.
//
// Returns:
//   - writer (*EventLog): the event log writer.
//   - err (error): a configuration or open error, or an error for a nil
//     handle.
func newEventLog(cfg *EventLogConfig, open eventLogOpen) (writer *EventLog, err error) {
	if cfg == nil || cfg.Source == "" {
		return nil, errors.New("writer: event log source is required")
	}

	eventID := cfg.EventID
	if eventID == 0 {
		eventID = 1
	}

	backend, err := open(cfg.Source)
	if err != nil {
		return nil, err
	}

	if backend == nil {
		return nil, errors.New("writer: event log open returned nil handle")
	}

	writer = &EventLog{
		backend: backend,
		eventID: eventID,
	}

	return writer, nil
}
