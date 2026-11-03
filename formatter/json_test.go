package formatter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
)

func TestJSONFormatter(t *testing.T) {
	t.Parallel()

	data, err := hqgologgerformatter.NewJSON(nil).Format(testLog())

	require.NoError(t, err)
	assert.True(t, json.Valid(data))
	assert.Contains(t, string(data), `"message":"msg"`)
	assert.Contains(t, string(data), `"level":"info"`)
	assert.Contains(t, string(data), `"k":"v"`)
	assert.Contains(t, string(data), `"n":7`)
}
