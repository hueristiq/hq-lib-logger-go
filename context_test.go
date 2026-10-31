package logger_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
)

func TestContextRoundTrip(t *testing.T) {
	t.Parallel()

	logger := hqgologger.New()
	ctx := hqgologger.NewContext(context.Background(), logger)

	stored, ok := hqgologger.FromContext(ctx)

	require.True(t, ok)
	assert.Same(t, logger, stored)
}

func TestContextNilLogger(t *testing.T) {
	t.Parallel()

	parent := context.Background()
	ctx := hqgologger.NewContext(parent, nil)

	assert.Equal(t, parent, ctx)

	stored, ok := hqgologger.FromContext(nil) //nolint:staticcheck // FromContext documents nil-context handling.

	assert.False(t, ok)
	assert.Nil(t, stored)
}

func TestContextWithoutLogger(t *testing.T) {
	t.Parallel()

	stored, ok := hqgologger.FromContext(context.Background())

	assert.False(t, ok)
	assert.Nil(t, stored)
}
