//nolint:testpackage // tests exercise unexported behavior.
package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextHelpers(t *testing.T) {
	t.Parallel()

	l := New()
	ctx := NewContext(context.Background(), l)

	got, ok := FromContext(ctx)

	require.True(t, ok)
	assert.Same(t, l, got)

	got, ok = FromContext(context.Background())
	assert.False(t, ok)
	assert.Nil(t, got)

	assert.Equal(t, context.Background(), NewContext(context.Background(), nil))
}

func TestFromContextNil(t *testing.T) {
	t.Parallel()

	//nolint:staticcheck // The nil guard in FromContext is deliberate; assert it reports ok=false.
	got, ok := FromContext(nil)

	assert.Nil(t, got)
	assert.False(t, ok)
}

func TestFromContextNonLoggerValue(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), contextKey{}, "not-a-logger")

	got, ok := FromContext(ctx)

	assert.Nil(t, got)
	assert.False(t, ok)
}
