package main

import (
	"time"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
)

func main() {
	logger := hqgologger.New(hqgologger.WithLevel(hqgologger.LevelDebug))
	defer logger.Close()

	worker := logger.Named("worker").With(
		hqgologger.String("component", "sync"),
		hqgologger.Duration("interval", 5*time.Millisecond),
	)

	worker.Info("ready", hqgologger.Int("jobs", 2))
}
