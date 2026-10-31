//nolint:testpackage // tests exercise unexported behavior.
package writer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorSentinels(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "writer: unsupported on this platform", ErrUnsupported.Error())
	assert.Equal(t, "writer: closed", ErrClosed.Error())

	assert.NotSame(t, ErrUnsupported, ErrClosed)

	wrappedUnsupported := fmt.Errorf("unsupported: %w", ErrUnsupported)
	wrappedClosed := fmt.Errorf("closed: %w", ErrClosed)

	require.NotErrorIs(t, wrappedClosed, ErrUnsupported)
	require.ErrorIs(t, wrappedUnsupported, ErrUnsupported)
	require.ErrorIs(t, wrappedClosed, ErrClosed)
}
