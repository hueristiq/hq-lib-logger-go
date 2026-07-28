# hq-lib-logger-go

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go reference](https://pkg.go.dev/badge/github.com/hueristiq/hq-lib-logger-go.svg)](https://pkg.go.dev/github.com/hueristiq/hq-lib-logger-go) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/blob/master/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-lib-logger-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-lib-logger-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/hq-lib-logger-go/blob/master/CONTRIBUTING.md)

`hq-lib-logger-go` is a [Go (Golang)](https://golang.org/) package for structured logging.

## Resources

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
	- [Quick start](#quick-start)
	- [Log levels](#log-levels)
	- [Attaching metadata](#attaching-metadata)
	- [Building a custom logger](#building-a-custom-logger)
	- [Colorized output](#colorized-output)
	- [Writing to multiple destinations](#writing-to-multiple-destinations)
- [Contributing](#contributing)
- [Licensing](#licensing)

## Features

- **Six severity levels**: `Fatal`, `Silent`, `Error`, `Info`, `Warn`, and `Debug`, with a configurable threshold.
- **Structured metadata**: attach typed key-value pairs to any message; they render as sorted `key=value` pairs.
- **Pluggable formatters**: the bundled console formatter handles timestamps, labels, and colorized labels; implement the `Formatter` interface for JSON, Logfmt, or anything else.
- **Flexible writers**: route logs to stdout, stderr, adapt any `io.Writer` with `IOWriter`, or fan out to several destinations at once with `MultiWriter`.
- **Optional color**: drop in the Fatih or Aurora colorizer, or stay plain with the no-op default.
- **Thread-safe**: the logger guards its configuration with a mutex and the console writer serializes its output.

## Installation

To install `hq-lib-logger-go`, run the following command in your Go project:

```bash
go get -v -u github.com/hueristiq/hq-lib-logger-go
```

## Usage

### Quick start

The package ships a `DefaultLogger`, pre-configured with a `LevelDebug` threshold, a console formatter (RFC3339 timestamps and labels), and a console writer that sends `Print` output to stdout and everything else to stderr. The package-level functions log through it:

```go
package main

import (
	"errors"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
)

func main() {
	hqgologger.Print("Application started", hqgologger.WithLabel("START"), hqgologger.WithString("app", "my-app"))
	hqgologger.Info("Processing request", hqgologger.WithString("request_id", "12345"))
	hqgologger.Warn("Resource usage high", hqgologger.WithString("memory", "80%"))
	hqgologger.Error("Failed to connect", hqgologger.WithError(errors.New("connection timeout")))
	hqgologger.Debug("Cache warmed", hqgologger.WithValue("entries", 42))
}
```

```
2025-08-08T13:45:00Z [START] Application started app=my-app
2025-08-08T13:45:00Z [INF] Processing request request_id=12345
2025-08-08T13:45:00Z [WRN] Resource usage high memory=80%
2025-08-08T13:45:00Z [ERR] Failed to connect

connection timeout
2025-08-08T13:45:00Z [DBG] Cache warmed entries=42
```

When a message carries no label, the console formatter supplies a default from the level: `FTL`, `ERR`, `INF`, `WRN`, or `DBG`. `Print` (level `Silent`) has no default label. An error attached with `WithError` prints as a trailing block, separated from the message by a blank line.

`Fatal` logs at the highest severity and then calls `os.Exit(1)` — which skips deferred functions — so place it only where you intend the program to stop.

### Log levels

Severity is ordered by value, and **lower means more severe**:

| Level | Value | Meaning |
| --- | --- | --- |
| `LevelFatal` | 0 | Unrecoverable failure; the logger exits the process after writing. |
| `LevelSilent` | 1 | As a threshold, suppresses output; as the level used by `Print`, routes to stdout. |
| `LevelError` | 2 | Errors that need attention but don't halt the program. |
| `LevelInfo` | 3 | Normal-operation messages. |
| `LevelWarn` | 4 | Potential issues worth investigating. |
| `LevelDebug` | 5 | Verbose detail for development. |

A logger emits an event only when it is at least as severe as the threshold — that is, when `event level <= configured level`. Setting the threshold to `LevelDebug` lets everything through; setting it to `LevelSilent` drops all but `Print`.

```go
import hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"

hqgologger.DefaultLogger.SetLevel(hqgologgerlevels.LevelWarn) // keeps Warn and anything more severe
```

Use `Enabled` to skip expensive message construction when a level would be discarded, and `Level` to read the current threshold:

```go
if hqgologger.DefaultLogger.Enabled(hqgologgerlevels.LevelDebug) {
	hqgologger.Debug(expensiveDump())
}
```

### Attaching metadata

Options configure a single log event:

| Option | Effect |
| --- | --- |
| `WithString(key, value)` | Adds a string metadata pair. |
| `WithValue(key, value)` | Adds a metadata pair of any type. |
| `WithError(err)` | Renders the error as a trailing block. |
| `WithLabel(label)` | Sets the bracketed label, overriding the default. |
| `WithoutLabel()` | Suppresses the label entirely. |
| `WithoutTimestamp()` | Drops the timestamp for this event. |

```go
hqgologger.Info("user signed in",
	hqgologger.WithString("user", "alex"),
	hqgologger.WithValue("attempts", 2),
)
// 2025-08-08T13:45:00Z [INF] user signed in attempts=2 user=alex
```

Metadata keys are sorted, so output stays stable across runs. The keys `label` and `error` are reserved — exported as `formatter.LabelKey` and `formatter.ErrorKey` — and are rendered specially rather than as `key=value` pairs.

### Building a custom logger

For full control, build a `Logger` and set its level, formatter, and writer yourself. A logger from `NewLogger` has none of these set and silently drops events until you configure them.

```go
package main

import (
	"errors"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

func main() {
	logger := hqgologger.NewLogger()

	if err := logger.SetLevel(hqgologgerlevels.LevelInfo); err != nil {
		panic(err)
	}

	logger.SetFormatter(hqgologgerformatter.NewConsoleFormatter(&hqgologgerformatter.ConsoleFormatterConfiguration{
		IncludeTimestamp: true,
		TimestampFormat:  "2006-01-02 15:04:05",
		IncludeLabel:     true,
	}))
	logger.SetWriter(hqgologgerwriter.NewConsoleWriter(&hqgologgerwriter.ConsoleWriterConfiguration{
		ForceStdout: true, // route every level to stdout
	}))

	logger.Info("Processing request", hqgologger.WithString("request_id", "67890"))
	logger.Error("Connection failed", hqgologger.WithError(errors.New("network error")))
}
```

Passing `nil` to `NewConsoleFormatter` or `NewConsoleWriter` applies the defaults from `DefaultConsoleFormatterConfig` and `DefaultConsoleWriterConfig`. When a logger's writer holds resources, release them with the logger's `Close` method.

### Colorized output

The default colorizer is a no-op, so labels print plain even with `Colorize: true`. For ANSI color, pick a colorizer from the `formatter/colorizer` package — one backed by [`fatih/color`](https://github.com/fatih/color), the other by [`logrusorgru/aurora`](https://github.com/logrusorgru/aurora):

```go
import (
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgercolorizer "github.com/hueristiq/hq-lib-logger-go/formatter/colorizer"
)

logger.SetFormatter(hqgologgerformatter.NewConsoleFormatter(&hqgologgerformatter.ConsoleFormatterConfiguration{
	IncludeLabel: true,
	Colorize:     true,
	Colorizer:    hqgologgercolorizer.NewFatihColorizer(),
}))
```

Each level maps to a distinct color; `Silent` is left uncolored.

### Writing to multiple destinations

`MultiWriter` forwards each message to every writer it wraps, skipping any `nil` entries:

```go
import hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"

logger.SetWriter(hqgologgerwriter.NewMultiWriter(
	hqgologgerwriter.NewConsoleWriter(nil),
	hqgologgerwriter.NewIOWriter(myFile), // adapts any io.Writer, no wrapper needed
))
```

`IOWriter` adapts a plain `io.Writer` (a file, a buffer, a network connection) to the `writer.Writer` interface, appending a newline to each message and closing the underlying writer on `Close` if it is closable. To add a custom destination with level-aware routing, implement the `writer.Writer` interface yourself: `Write(data []byte, level levels.Level) error` and `Close() error`.

## Contributing

Contributions are welcome and encouraged! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-lib-logger-go/pulls) or report [Issues](https://github.com/hueristiq/hq-lib-logger-go/issues). For more details, check out the [contribution guidelines](https://github.com/hueristiq/hq-lib-logger-go/blob/main/CONTRIBUTING.md).

A big thank you to all the [contributors](https://github.com/hueristiq/hq-lib-logger-go/graphs/contributors) for your ongoing support!

![contributors](https://contrib.rocks/image?repo=hueristiq/hq-lib-logger-go&max=500)

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-lib-logger-go/blob/main/LICENSE).
