package main

import (
	"errors"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
)

func main() {
	err := errors.New("root error example!")

	hqgologger.Error("Error message", hqgologger.WithString("string-key", "string-value"), hqgologger.WithValue("value-key", "value-value"), hqgologger.WithError(err))
}
