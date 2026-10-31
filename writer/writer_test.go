package writer_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

func TestStream(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	stream := hqgologgerwriter.NewStream(&buf)

	require.NoError(t, stream.Write([]byte("record"), hqgologgerlevel.LevelInfo))
	assert.Equal(t, "record\n", buf.String())

	require.NoError(t, stream.Close())
	require.NoError(t, stream.Close())

	assert.ErrorIs(t, stream.Write([]byte("after close"), hqgologgerlevel.LevelInfo), hqgologgerwriter.ErrClosed)
}

func TestMultiWriter(t *testing.T) {
	t.Parallel()

	var first, second bytes.Buffer

	multi := hqgologgerwriter.NewMulti(
		hqgologgerwriter.NewStream(&first),
		nil,
		hqgologgerwriter.NewStream(&second),
	)

	require.NoError(t, multi.Write([]byte("record"), hqgologgerlevel.LevelInfo))
	assert.Equal(t, "record\n", first.String())
	assert.Equal(t, "record\n", second.String())

	require.NoError(t, multi.Close())
	assert.ErrorIs(t, multi.Write([]byte("after close"), hqgologgerlevel.LevelInfo), hqgologgerwriter.ErrClosed)
}

func TestNilStreamDestination(t *testing.T) {
	t.Parallel()

	stream := hqgologgerwriter.NewStream(nil)

	require.NoError(t, stream.Write([]byte("record"), hqgologgerlevel.LevelInfo))
	require.NoError(t, stream.Close())
}
