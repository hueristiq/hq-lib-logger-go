//nolint:testpackage // tests exercise unexported behavior.
package writer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

type fakeSyslogBackend struct {
	priorities []int
	messages   []string
	closeErr   error
	closed     bool
}

func (f *fakeSyslogBackend) WriteLevel(priority int, message string) error {
	f.priorities = append(f.priorities, priority)
	f.messages = append(f.messages, message)

	return nil
}

func (f *fakeSyslogBackend) Close() error {
	f.closed = true

	return f.closeErr
}

func TestSyslogWritesPriorities(t *testing.T) {
	t.Parallel()

	backend := &fakeSyslogBackend{}
	writer := &Syslog{backend: backend, facility: syslogFacilityUser}

	require.NoError(t, writer.Write([]byte("fatal"), hqgologgerlevel.LevelFatal))
	require.NoError(t, writer.Write([]byte("error"), hqgologgerlevel.LevelError))
	require.NoError(t, writer.Write([]byte("warn"), hqgologgerlevel.LevelWarn))
	require.NoError(t, writer.Write([]byte("info"), hqgologgerlevel.LevelInfo))
	require.NoError(t, writer.Write([]byte("debug"), hqgologgerlevel.LevelDebug))
	require.NoError(t, writer.Write([]byte("trace"), hqgologgerlevel.LevelTrace))

	assert.Equal(t, []int{
		syslogFacilityUser | syslogSeverityCritical,
		syslogFacilityUser | syslogSeverityError,
		syslogFacilityUser | syslogSeverityWarning,
		syslogFacilityUser | syslogSeverityInfo,
		syslogFacilityUser | syslogSeverityDebug,
		syslogFacilityUser | syslogSeverityDebug,
	}, backend.priorities)
	assert.Equal(t, []string{"fatal", "error", "warn", "info", "debug", "trace"}, backend.messages)
}

func TestSyslogCloseAndInvalidLevel(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	backend := &fakeSyslogBackend{closeErr: closeErr}
	writer := &Syslog{backend: backend}

	require.ErrorIs(t, writer.Write([]byte("x"), hqgologgerlevel.Level(99)), hqgologgerlevel.ErrUnknownLevel)
	require.ErrorIs(t, writer.Close(), closeErr)
	assert.True(t, backend.closed)
	require.ErrorIs(t, writer.Close(), closeErr)
	require.ErrorIs(t, writer.Write([]byte("x"), hqgologgerlevel.LevelInfo), ErrClosed)
}

func TestSyslogUnsupportedNilBackend(t *testing.T) {
	t.Parallel()

	var writer *Syslog

	require.ErrorIs(t, writer.Write(nil, hqgologgerlevel.LevelInfo), ErrUnsupported)
	require.NoError(t, writer.Close())
}

func TestSyslogSeverityInvalid(t *testing.T) {
	t.Parallel()

	severity, ok := syslogSeverity(hqgologgerlevel.Level(99))

	assert.Zero(t, severity)
	assert.False(t, ok)
}

func TestSyslogZeroFacilityDefaults(t *testing.T) {
	t.Parallel()

	backend := &fakeSyslogBackend{}

	writer := &Syslog{backend: backend}

	require.NoError(t, writer.Write([]byte("x"), hqgologgerlevel.LevelInfo))
}
