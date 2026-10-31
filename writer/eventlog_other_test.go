//go:build !windows

//nolint:testpackage // tests exercise unexported behavior.
package writer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEventLogUnsupported(t *testing.T) {
	t.Parallel()

	_, err := NewEventLog(nil)
	require.ErrorIs(t, err, ErrUnsupported)

	_, err = NewEventLog(&EventLogConfig{Source: "app"})
	require.ErrorIs(t, err, ErrUnsupported)
}
