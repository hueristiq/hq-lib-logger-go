//go:build !windows && !plan9

//nolint:testpackage // tests exercise unexported behavior.
package writer

import (
	"errors"
	"log/syslog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSyslogLevelWriter struct {
	calls []string
}

func (f *fakeSyslogLevelWriter) Crit(string) error {
	f.calls = append(f.calls, "crit")

	return nil
}

func (f *fakeSyslogLevelWriter) Err(string) error {
	f.calls = append(f.calls, "err")

	return nil
}

func (f *fakeSyslogLevelWriter) Warning(string) error {
	f.calls = append(f.calls, "warning")

	return nil
}

func (f *fakeSyslogLevelWriter) Notice(string) error {
	f.calls = append(f.calls, "notice")

	return nil
}

func (f *fakeSyslogLevelWriter) Info(string) error {
	f.calls = append(f.calls, "info")

	return nil
}

func (f *fakeSyslogLevelWriter) Debug(string) error {
	f.calls = append(f.calls, "debug")

	return nil
}

func (f *fakeSyslogLevelWriter) Close() error {
	return nil
}

func TestSyslogBackendAdapter(t *testing.T) {
	t.Parallel()

	writer := &fakeSyslogLevelWriter{}
	adapter := &syslogBackendAdapter{writer: writer}

	require.NoError(t, adapter.WriteLevel(syslogSeverityCritical, "x"))
	require.NoError(t, adapter.WriteLevel(syslogSeverityError, "x"))
	require.NoError(t, adapter.WriteLevel(syslogSeverityWarning, "x"))
	require.NoError(t, adapter.WriteLevel(syslogSeverityNotice, "x"))
	require.NoError(t, adapter.WriteLevel(syslogSeverityInfo, "x"))
	require.NoError(t, adapter.WriteLevel(syslogSeverityDebug, "x"))
	require.ErrorIs(t, adapter.WriteLevel(1, "x"), ErrUnsupported)

	assert.Equal(t, []string{"crit", "err", "warning", "notice", "info", "debug"}, writer.calls)
}

func TestNewSyslogDefaults(t *testing.T) {
	t.Parallel()

	expected := errors.New("stop")

	var (
		network  string
		address  string
		priority syslog.Priority
		tag      string
	)

	dial := func(n, a string, p syslog.Priority, t string) (*syslog.Writer, error) {
		network, address, priority, tag = n, a, p, t

		return nil, expected
	}

	_, err := newSyslog(&SyslogConfig{Network: "udp", Address: "localhost:514", Facility: 3 << 3}, dial)

	require.ErrorIs(t, err, expected)
	assert.Equal(t, "udp", network)
	assert.Equal(t, "localhost:514", address)
	assert.Equal(t, syslog.Priority(3<<3), priority)
	assert.Equal(t, "hq-lib-logger-go", tag)
}

func TestNewSyslogNilDialResult(t *testing.T) {
	t.Parallel()

	dial := func(string, string, syslog.Priority, string) (*syslog.Writer, error) {
		//nolint:nilnil // The constructor must reject a nil connection.
		return nil, nil
	}

	_, err := newSyslog(nil, dial)

	require.Error(t, err)
}

func TestNewSyslogRealDialError(t *testing.T) {
	t.Parallel()

	_, err := NewSyslog(&SyslogConfig{Network: "tcp", Address: "127.0.0.1:1"})

	require.Error(t, err)
}

func TestSyslogAdapterClose(t *testing.T) {
	t.Parallel()

	adapter := &syslogBackendAdapter{writer: &fakeSyslogLevelWriter{}}

	require.NoError(t, adapter.Close())
}

func TestNewSyslogSuccess(t *testing.T) {
	t.Parallel()

	dial := func(string, string, syslog.Priority, string) (*syslog.Writer, error) {
		return &syslog.Writer{}, nil
	}

	writer, err := newSyslog(nil, dial)

	require.NoError(t, err)
	require.NotNil(t, writer)
	assert.Equal(t, syslogFacilityUser, writer.facility)
}
