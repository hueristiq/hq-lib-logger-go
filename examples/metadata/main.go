package main

import (
	hqgologger "github.com/hueristiq/hq-lib-logger-go"
)

func main() {
	hqgologger.Print("Print message", hqgologger.WithLabel("PRINT"), hqgologger.WithString("string-key", "string-value"), hqgologger.WithValue("value-key", "value-value"))
	hqgologger.Info("Info message", hqgologger.WithLabel("INFO"), hqgologger.WithString("string-key", "string-value"), hqgologger.WithValue("value-key", "value-value"))
	hqgologger.Warn("Warn message", hqgologger.WithLabel("WARN"), hqgologger.WithString("string-key", "string-value"), hqgologger.WithValue("value-key", "value-value"))
	hqgologger.Error("Error message", hqgologger.WithLabel("ERROR"), hqgologger.WithString("string-key", "string-value"), hqgologger.WithValue("value-key", "value-value"))
	hqgologger.Fatal("Fatal message", hqgologger.WithLabel("FATAL"), hqgologger.WithString("string-key", "string-value"), hqgologger.WithValue("value-key", "value-value"))
}
