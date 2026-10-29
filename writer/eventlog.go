package writer

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hueristiq/hq-lib-logger-go/levels"
)

// EventLogConfig configures a Windows event log writer.
//
// Fields:
//   - Source (string): registered event source name; required.
//   - EventID (uint32): event identifier; zero selects 1.
type EventLogConfig struct {
	Source  string
	EventID uint32
}

// eventLogBackend abstracts the Windows event log connection, keeping the
// writer testable.
//
// Methods:
//   - Info(eventID uint32, message string) (err error): writes an
//     informational event.
//   - Warning(eventID uint32, message string) (err error): writes a warning
//     event.
//   - Error(eventID uint32, message string) (err error): writes an error
//     event.
//   - Close() (err error): releases the handle.
type eventLogBackend interface {
	Info(eventID uint32, message string) error
	Warning(eventID uint32, message string) error
	Error(eventID uint32, message string) error
	Close() error
}

// EventLog writes records to the Windows event log. The constructor returns
// [ErrUnsupported] on other platforms.
//
// Fields:
//   - mutex (sync.Mutex): serializes writes and Close.
//   - backend (eventLogBackend): platform event log handle.
//   - eventID (uint32): event identifier.
//   - closed (bool): whether Close has run.
//   - closeErr (error): result of the first Close.
type EventLog struct {
	mutex    sync.Mutex
	backend  eventLogBackend
	eventID  uint32
	closed   bool
	closeErr error
}

// Write sends data as an event whose type follows level.
//
// Parameters:
//   - data ([]byte): formatted record.
//   - level (levels.Level): record severity.
//
// Returns:
//   - err (error): [ErrUnsupported], [ErrClosed], or a backend error.
func (e *EventLog) Write(data []byte, level levels.Level) (err error) {
	if e == nil || e.backend == nil {
		return ErrUnsupported
	}

	if !level.Valid() {
		return fmt.Errorf("%w (%d)", levels.ErrUnknownLevel, int(level))
	}

	message := strings.TrimRight(string(data), "\r\n")

	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.closed {
		return ErrClosed
	}

	switch level {
	case levels.LevelFatal, levels.LevelError:
		return e.backend.Error(e.eventID, message)
	case levels.LevelWarn:
		return e.backend.Warning(e.eventID, message)
	case levels.LevelInfo, levels.LevelDebug, levels.LevelTrace:
		return e.backend.Info(e.eventID, message)
	default:
		// Unreachable: the Valid check above admits only the six levels,
		// all of which are listed here. Kept defensively.
		return e.backend.Info(e.eventID, message)
	}
}

// Close closes the event log handle. Repeated calls return the first error.
//
// Returns:
//   - err (error): the backend's close error, if any.
func (e *EventLog) Close() (err error) {
	if e == nil || e.backend == nil {
		return nil
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.closed {
		return e.closeErr
	}

	e.closed = true
	e.closeErr = e.backend.Close()

	return e.closeErr
}

var _ Writer = (*EventLog)(nil)
