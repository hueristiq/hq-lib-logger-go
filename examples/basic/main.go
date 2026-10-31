package main

import (
	hqgologger "github.com/hueristiq/hq-lib-logger-go"
)

func main() {
	logger := hqgologger.New(hqgologger.WithLevel(hqgologger.LevelDebug))
	defer logger.Close()

	logger.Info("service started",
		hqgologger.String("service", "example"),
		hqgologger.Int("attempt", 1),
	)
}
