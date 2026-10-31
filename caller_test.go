//nolint:testpackage // tests exercise unexported behavior.
package logger

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaptureCaller(t *testing.T) {
	t.Parallel()

	caller := captureCaller()

	require.NotEmpty(t, caller)
	assert.Contains(t, caller, "caller_test.go:")
}

func TestCallerFromPC(t *testing.T) {
	t.Parallel()

	assert.Empty(t, callerFromPC(0))
}

func TestIsInternalFrame(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		function string
		file     string
		want     bool
	}{
		{"logger source", modulePath + ".Logger.Info", "logger.go", true},
		{"logger test source", modulePath + ".TestCaptureCaller", "caller_test.go", false},
		{"external source", "main.main", "main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			frame := &runtime.Frame{Function: tt.function, File: tt.file}

			assert.Equal(t, tt.want, isInternalFrame(frame))
		})
	}
}

func TestIsWrapperFrame(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		function string
		want     bool
	}{
		{"fmt wrapper", "fmt.Println", true},
		{"io wrapper", "io.Copy", true},
		{"slog wrapper", "log/slog.(*Logger).Info", true},
		{"ordinary function", "main.main", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			frame := &runtime.Frame{Function: tt.function}

			assert.Equal(t, tt.want, isWrapperFrame(frame))
		})
	}
}

func TestShortCaller(t *testing.T) {
	t.Parallel()

	frame := &runtime.Frame{File: "/home/user/project/cmd/app/main.go", Line: 42}

	assert.Equal(t, "main.go:42", shortCaller(frame))
}
