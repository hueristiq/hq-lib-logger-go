//go:build windows || plan9

package writer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSyslogUnsupported(t *testing.T) {
	t.Parallel()

	_, err := NewSyslog(nil)

	require.ErrorIs(t, err, ErrUnsupported)
}
