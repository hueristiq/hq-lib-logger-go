package logger_test

import (
	"io"
	"testing"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

func benchmarkEmit(b *testing.B, f hqgologgerformatter.Formatter) {
	b.Helper()

	l := hqgologger.New(
		hqgologger.WithFormatter(f),
		hqgologger.WithWriter(hqgologgerwriter.NewStream(io.Discard)),
		hqgologger.WithLevel(hqgologger.LevelTrace),
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := range b.N {
		l.Info("message", hqgologger.String("key", "value"), hqgologger.Int("n", i))
	}
}

func BenchmarkEmitConsole(b *testing.B) {
	benchmarkEmit(b, hqgologgerformatter.NewConsole(&hqgologgerformatter.ConsoleConfig{Timestamp: false, Label: true}))
}

func BenchmarkEmitJSON(b *testing.B) {
	benchmarkEmit(b, hqgologgerformatter.NewJSON(nil))
}
