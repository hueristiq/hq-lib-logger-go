package formatter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hueristiq/hq-lib-logger-go/levels"
)

func TestLogfmtDefaults(t *testing.T) {
	t.Parallel()

	data, err := NewLogfmt(nil).Format(&Log{
		Level:   levels.LevelInfo,
		Message: "hello world",
		Fields:  []Field{String("b", "2"), String("a", "1")},
	})

	require.NoError(t, err)
	assert.Equal(t, `level=info msg="hello world" a=1 b=2`, string(data))
}

func TestLogfmtCustomKeysAndError(t *testing.T) {
	t.Parallel()

	formatter := NewLogfmt(&LogfmtConfig{
		LevelKey:   "lvl",
		MessageKey: "msg",
		ErrorKey:   "err",
	})

	data, err := formatter.Format(&Log{
		Level:   levels.LevelWarn,
		Message: "warned",
		Fields: []Field{
			String("k", "v"),
			Err("err", errors.New("boom")),
		},
	})

	require.NoError(t, err)
	assert.Equal(t, `lvl=warn msg=warned err=boom k=v`, string(data))
}

func TestLogfmtQuotesUnsafeValues(t *testing.T) {
	t.Parallel()

	data, err := NewLogfmt(&LogfmtConfig{}).Format(&Log{
		Level:   levels.LevelInfo,
		Message: "line\nbreak",
		Fields:  []Field{String("path", `a=b`), String("empty", "")},
	})

	require.NoError(t, err)
	assert.Contains(t, string(data), `msg="line\nbreak"`)
	assert.Contains(t, string(data), `empty=`)
	assert.Contains(t, string(data), `path="a=b"`)
}

func TestLogfmtInvalidLevel(t *testing.T) {
	t.Parallel()

	_, err := NewLogfmt(nil).Format(&Log{Level: levels.Level(99)})

	require.ErrorIs(t, err, levels.ErrUnknownLevel)
}
