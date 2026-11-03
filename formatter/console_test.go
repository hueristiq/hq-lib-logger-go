package formatter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
)

func TestConsoleFormatter(t *testing.T) {
	t.Parallel()

	data, err := hqgologgerformatter.NewConsole(nil).Format(testLog())

	require.NoError(t, err)
	assert.Contains(t, string(data), "msg")
	assert.Contains(t, string(data), "k=v")
}
