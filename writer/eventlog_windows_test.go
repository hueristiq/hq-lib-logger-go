//go:build windows

package writer

import (
	"testing"

	"golang.org/x/sys/windows/svc/eventlog"

	"github.com/stretchr/testify/require"
)

func TestNewEventLogRequiresSource(t *testing.T) {
	t.Parallel()

	open := func(string) (*eventlog.Log, error) {
		t.Fatal("open must not be called")

		return nil, nil
	}

	_, err := newEventLog(nil, open)
	require.Error(t, err)

	_, err = newEventLog(&EventLogConfig{}, open)
	require.Error(t, err)
}
