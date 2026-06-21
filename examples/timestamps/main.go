package main

import (
	hqgologger "github.com/hueristiq/hq-lib-logger-go"
)

func main() {
	hqgologger.Print("Print message with timestamp", hqgologger.WithLabel("PRINT"))
	hqgologger.Print("Print message without timestamp", hqgologger.WithLabel("PRINT"), hqgologger.WithoutTimestamp())
}
