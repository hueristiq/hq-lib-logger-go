package logger

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
)

func reconfigureDefault(t *testing.T) *captureWriter {
	t.Helper()

	prevLogger := DefaultLogger

	t.Cleanup(func() { DefaultLogger = prevLogger })

	w := &captureWriter{}

	l := NewLogger()

	_ = l.SetLevel(hqgologgerlevels.LevelDebug) // LevelDebug is valid; cannot fail.

	l.SetFormatter(hqgologgerformatter.NewConsoleFormatter(&hqgologgerformatter.ConsoleFormatterConfiguration{
		IncludeLabel: true,
		Colorizer:    hqgologgerformatter.NewNoOpColorizer(),
	}))
	l.SetWriter(w)

	DefaultLogger = l

	return w
}

//nolint:paralleltest // mutates package-global DefaultLogger
func TestDefaultLoggerInitialized(t *testing.T) {
	require.NotNil(t, DefaultLogger)
}

//nolint:paralleltest // mutates package-global DefaultLogger
func TestPackageInfo(t *testing.T) {
	w := reconfigureDefault(t)

	Info("informational")
	assert.Contains(t, w.String(), "[INF] informational")
}

//nolint:paralleltest // mutates package-global DefaultLogger
func TestPackageWarn(t *testing.T) {
	w := reconfigureDefault(t)

	Warn("warning")
	assert.Contains(t, w.String(), "[WRN] warning")
}

//nolint:paralleltest // mutates package-global DefaultLogger
func TestPackageError(t *testing.T) {
	w := reconfigureDefault(t)

	Error("erroneous")
	assert.Contains(t, w.String(), "[ERR] erroneous")
}

//nolint:paralleltest // mutates package-global DefaultLogger
func TestPackageDebug(t *testing.T) {
	w := reconfigureDefault(t)

	Debug("debugging")
	assert.Contains(t, w.String(), "[DBG] debugging")
}

//nolint:paralleltest // mutates package-global DefaultLogger
func TestPackagePrint(t *testing.T) {
	w := reconfigureDefault(t)

	Print("printed")

	out := w.String()
	assert.Contains(t, out, "printed")
	assert.NotContains(t, out, "[")
}

//nolint:paralleltest // mutates package-global DefaultLogger
func TestPackageInfoWithOptions(t *testing.T) {
	w := reconfigureDefault(t)

	Info("msg", WithString("k", "v"), WithLabel("L"))

	out := w.String()
	assert.Contains(t, out, "[L] msg")
	assert.Contains(t, out, "k=v")
}

//nolint:paralleltest // mutates package-global DefaultLogger
func TestPackageConcurrentLogging(t *testing.T) {
	w := reconfigureDefault(t)

	var wg sync.WaitGroup

	for range 16 {
		wg.Go(func() {
			Info("x")
		})
	}

	wg.Wait()

	assert.Equal(t, 16, strings.Count(w.String(), "[INF] x"))
}
