package logger

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// modulePath prefixes every function in this module. Frames from this module
// are skipped when resolving the caller, unless they come from a _test.go file
// (which is where a caller in this package's own tests lives).
const modulePath = "github.com/hueristiq/hq-lib-logger-go"

// captureCaller returns the first stack frame outside the logger and its
// conventional wrappers, formatted as "file.go:line".
//
// Returns:
//   - caller (string): short caller location, or "" when none is found.
func captureCaller() (caller string) {
	var pcs [16]uintptr

	n := runtime.Callers(2, pcs[:])

	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()

		if !isInternalFrame(&frame) && !isWrapperFrame(&frame) {
			return shortCaller(&frame)
		}

		// The no-frame fallback is defensive: any goroutine has at least one
		// frame, so more is always true in practice.
		if !more {
			return ""
		}
	}
}

// callerFromPC formats the caller recorded in a slog.Record.
//
// Parameters:
//   - pc (uintptr): program counter from slog.Record.PC.
//
// Returns:
//   - caller (string): short caller location, or "" when pc is zero.
func callerFromPC(pc uintptr) (caller string) {
	if pc == 0 {
		return ""
	}

	frames := runtime.CallersFrames([]uintptr{pc})

	frame, _ := frames.Next()

	return shortCaller(&frame)
}

// isInternalFrame reports whether frame belongs to this module, excluding the
// module's own tests, whose frames represent the caller.
//
// Parameters:
//   - frame (runtime.Frame): candidate frame.
//
// Returns:
//   - internal (bool): true when the frame should be skipped.
func isInternalFrame(frame *runtime.Frame) (internal bool) {
	if !strings.HasPrefix(frame.Function, modulePath) {
		return false
	}

	return !strings.HasSuffix(frame.File, "_test.go")
}

// isWrapperFrame reports whether frame belongs to a standard-library package
// that commonly sits between the caller and the logger (for example when
// logging through Logger.Writer).
//
// Parameters:
//   - frame (runtime.Frame): candidate frame.
//
// Returns:
//   - wrapper (bool): true when the frame should be skipped.
func isWrapperFrame(frame *runtime.Frame) (wrapper bool) {
	return strings.HasPrefix(frame.Function, "fmt.") ||
		strings.HasPrefix(frame.Function, "io.") ||
		strings.HasPrefix(frame.Function, "log/slog.")
}

// shortCaller formats a frame as "file.go:line".
//
// Parameters:
//   - frame (runtime.Frame): the frame to format.
//
// Returns:
//   - caller (string): short caller location.
func shortCaller(frame *runtime.Frame) (caller string) {
	return filepath.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
}
