//nolint:testpackage // tests exercise unexported behavior.
package writer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

type fakeEventLogBackend struct {
	infos    int
	warnings int
	errors   int
	closed   bool
}

func (f *fakeEventLogBackend) Info(uint32, string) error {
	f.infos++

	return nil
}

func (f *fakeEventLogBackend) Warning(uint32, string) error {
	f.warnings++

	return nil
}

func (f *fakeEventLogBackend) Error(uint32, string) error {
	f.errors++

	return nil
}

func (f *fakeEventLogBackend) Close() error {
	f.closed = true

	return nil
}

func TestEventLogMapping(t *testing.T) {
	t.Parallel()

	backend := &fakeEventLogBackend{}
	writer := &EventLog{backend: backend, eventID: 7}

	require.NoError(t, writer.Write([]byte("fatal"), hqgologgerlevel.LevelFatal))
	require.NoError(t, writer.Write([]byte("error"), hqgologgerlevel.LevelError))
	require.NoError(t, writer.Write([]byte("warn"), hqgologgerlevel.LevelWarn))
	require.NoError(t, writer.Write([]byte("info"), hqgologgerlevel.LevelInfo))

	assert.Equal(t, 2, backend.errors)
	assert.Equal(t, 1, backend.warnings)
	assert.Equal(t, 1, backend.infos)
}

func TestEventLogCloseAndInvalidLevel(t *testing.T) {
	t.Parallel()

	backend := &fakeEventLogBackend{}
	writer := &EventLog{backend: backend}

	require.ErrorIs(t, writer.Write([]byte("x"), hqgologgerlevel.Level(99)), hqgologgerlevel.ErrUnknownLevel)
	require.NoError(t, writer.Close())
	assert.True(t, backend.closed)
	require.ErrorIs(t, writer.Write([]byte("x"), hqgologgerlevel.LevelInfo), ErrClosed)
}

func TestEventLogNilBackend(t *testing.T) {
	t.Parallel()

	var writer *EventLog

	require.ErrorIs(t, writer.Write(nil, hqgologgerlevel.LevelInfo), ErrUnsupported)
	require.NoError(t, writer.Close())
}

func TestEventLogZeroValueClose(t *testing.T) {
	t.Parallel()

	zero := &EventLog{}

	require.NoError(t, zero.Close())
}

func TestEventLogSecondClose(t *testing.T) {
	t.Parallel()

	backend := &fakeEventLogBackend{}
	writer := &EventLog{backend: backend}

	require.NoError(t, writer.Close())
	require.NoError(t, writer.Close())
	assert.True(t, backend.closed)
}
