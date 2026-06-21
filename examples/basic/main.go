package main

import (
	hqgologger "github.com/hueristiq/hq-lib-logger-go"
)

func main() {
	hqgologger.Print("Print message")
	hqgologger.Info("Info message")
	hqgologger.Warn("Warn message")
	hqgologger.Error("Error message")
	hqgologger.Fatal("Fatal message")
}
