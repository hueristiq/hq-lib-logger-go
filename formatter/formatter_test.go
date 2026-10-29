package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundledFormattersImplementInterfaces(t *testing.T) {
	t.Parallel()

	var (
		_ Formatter = NewConsole(nil)
		_ Appender  = NewConsole(nil)
		_ Formatter = NewJSON(nil)
		_ Appender  = NewJSON(nil)
		_ Formatter = NewLogfmt(nil)
		_ Appender  = NewLogfmt(nil)
	)
}

func TestFormattersRejectNilLog(t *testing.T) {
	t.Parallel()

	formatters := []Formatter{
		NewConsole(nil),
		NewJSON(nil),
		NewLogfmt(nil),
	}

	for _, formatter := range formatters {
		data, err := formatter.Format(nil)

		assert.Nil(t, data)
		require.ErrorIs(t, err, ErrNilLog)
	}
}

func TestReservedKeys(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "label", LabelKey)
	assert.Equal(t, "error", ErrorKey)
	assert.Equal(t, "caller", CallerKey)
}
