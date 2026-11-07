# hq-lib-logger-go

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go reference](https://pkg.go.dev/badge/github.com/hueristiq/hq-lib-logger-go.svg)](https://pkg.go.dev/github.com/hueristiq/hq-lib-logger-go) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/blob/main/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-lib-logger-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-lib-logger-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/issues?q=is:issue+is:closed)

`hq-lib-logger-go` is a [Go](https://golang.org/) package for structured logging.

## Resources

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
	- [Quick start](#quick-start)
	- [Log levels](#log-levels)
	- [Structured fields](#structured-fields)
	- [Child loggers](#child-loggers)
	- [Formatters](#formatters)
	- [Writers](#writers)
	- [log/slog integration](#logslog-integration)
	- [Global logger](#global-logger)
	- [Building a logger](#building-a-logger)
	- [Custom formatters and writers](#custom-formatters-and-writers)
- [Contributing](#contributing)
- [Licensing](#licensing)

## Features

- **Levels and fields:** `Fatal`, `Error`, `Warn`, `Info`, `Debug`, and `Trace` sit behind a configurable threshold, with typed fields such as `String`, `Int`, `Duration`, `Time`, and `Err`.
- **Child views:** `Named`, `With`, and `WithPrefix` share one configuration core.
- **Output:** console, JSON, and logfmt formatters; stream, fan-out, syslog, and Windows event log writers.
- **slog integration:** `SlogHandler` maps slog levels, attributes, and groups to logger fields and never exits for high levels.
- **Runtime control:** `SetLevel` changes the threshold, and `LogRecord` accepts a full record with a custom time, caller, and name.
- **Standard-library runtime:** the optional Windows event log writer uses `golang.org/x/sys` on Windows.

## Installation

To install `hq-lib-logger-go`, run the following command in your Go project:

```bash
go get github.com/hueristiq/hq-lib-logger-go@latest
```

Requires Go 1.26.0 or later.

## Usage

The examples import the package under the `hqgologger` alias. Subpackages use `hqgologgerformatter`, `hqgologgerlevel`, and `hqgologgerwriter`.

### Quick start

`New()` is usable immediately. It uses a console formatter, writes to stderr, and starts at `LevelInfo`.

```go
package main

import (
	"log/slog"
	"os"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)

func main() {
	log := hqgologger.New(
		hqgologger.WithLevel(hqgologger.LevelDebug),
		hqgologger.WithFormatter(hqgologgerformatter.NewJSON(nil)),
		hqgologger.WithWriter(hqgologgerwriter.NewStream(os.Stdout)),
	)

	log.Info("application started", hqgologger.String("service", "api"))

	api := log.Named("api").With(hqgologger.String("request_id", "42"))
	api.Info("request finished", hqgologger.Int("status", 200))

	slog.SetDefault(slog.New(log.SlogHandler()))
	slog.Info("hello", "request_id", "42")
}
```

### Log levels

| Level | Value | Meaning |
| --- | --- | --- |
| `LevelFatal` | 0 | Unrecoverable failure; logged like any other level. |
| `LevelError` | 1 | Errors that need attention. |
| `LevelWarn` | 2 | Potential issues worth investigating. |
| `LevelInfo` | 3 | Normal-operation messages. |
| `LevelDebug` | 4 | Verbose detail for development. |
| `LevelTrace` | 5 | The most verbose output. |

An event is emitted only when `event level <= configured level`:

```go
_ = log.SetLevel(hqgologger.LevelDebug)

if log.Enabled(hqgologger.LevelTrace) {
	// Trace is below the current threshold.
}
```

All levels, including `Fatal`, are subject to sampling.

### Structured fields

All level methods and `Log` accept typed fields:

```go
log.Info("request finished",
	hqgologger.String("method", "GET"),
	hqgologger.String("path", "/healthz"),
	hqgologger.Int("status", 200),
	hqgologger.Duration("latency", 3*time.Millisecond),
	hqgologger.Bool("cached", true),
	hqgologger.Time("at", time.Now()),
	hqgologger.Err("cause", err),
	hqgologger.Any("headers", []string{"accept: */*"}),
)
```

Use `Log` for an explicit level:

```go
log.Log(hqgologger.LevelWarn, "slow response", hqgologger.Duration("latency", latency))
```

### Child loggers

`With` attaches fields, `Named` adds a dotted logger name, and `WithPrefix` prepends raw text to messages:

```go
api := log.Named("api").With(hqgologger.String("service", "api"))

api.Info("started")
api.WithPrefix("[request] ").Info("finished")
```

Child views share the logger's level, formatter, and writer. Fields, name, and prefix belong to the view.

### Formatters

```go
hqgologgerformatter.NewConsole(nil) // timestamp, label, sorted fields
hqgologgerformatter.NewJSON(nil)    // one JSON object per record
hqgologgerformatter.NewLogfmt(nil)  // compact key=value
```

Each constructor accepts a nil config for package defaults. A non-nil zero config is explicit: timestamps and labels are off, and optional keys are omitted. This JSON formatter emits no timestamp and nests fields:

```go
formatter := hqgologgerformatter.NewJSON(&hqgologgerformatter.JSONConfig{
	FieldsKey: "fields",
})
```

### Writers

```go
hqgologgerwriter.NewStream(os.Stdout) // one destination, one newline per record
hqgologgerwriter.NewMulti(w1, w2)     // fan out to several writers
hqgologgerwriter.NewSyslog(nil)       // Unix syslog
hqgologgerwriter.NewEventLog(cfg)     // Windows event log
```

`Stream` serializes writes and flushes when the destination implements `Flush`. `MultiWriter` attempts every destination and joins errors. Platform writers return `hqgologgerwriter.ErrUnsupported` where unavailable.

For file output with rotation, compose `Stream` with any rotating `io.Writer`, such as [lumberjack](https://github.com/natefinch/lumberjack):

```go
log := hqgologger.New(hqgologger.WithWriter(
	hqgologgerwriter.NewStream(&lumberjack.Logger{
		Filename:   "app.log",
		MaxSize:    100, // megabytes
		MaxBackups: 3,
	}),
))
```

### log/slog integration

```go
slog.SetDefault(slog.New(log.SlogHandler()))

slog.Info("hello", "request_id", "42")
```

The handler maps slog levels to logger levels, converts attributes and groups to typed fields, and never exits for high slog levels.

### Global logger

The package intentionally has no mutable package-level logger. For `slog.Info`-style convenience, install a default handler once at startup:

```go
slog.SetDefault(slog.New(log.SlogHandler()))

slog.Info("hello", "request_id", "42") // package-level, goes through log
```

### Building a logger

```go
log := hqgologger.New(
	hqgologger.WithLevel(hqgologgerlevel.LevelInfo),
	hqgologger.WithFormatter(hqgologgerformatter.NewJSON(nil)),
	hqgologger.WithWriter(hqgologgerwriter.NewStream(os.Stdout)),
	hqgologger.WithCallerCapture(true),
)
```

Options are applied in order. Nil formatters and writers are ignored. An invalid level is ignored and the default `LevelInfo` remains. `SetLevel` is the only runtime configuration mutator; create a new logger to change the formatter, writer, or caller capture.

`LogRecord` is the advanced path for a full `hqgologgerformatter.Log`, including a custom time, caller, and name:

```go
log.LogRecord(hqgologgerformatter.Log{
	Level:   hqgologger.LevelInfo,
	Message: "import finished",
	Fields:  []hqgologger.Field{hqgologger.Int("rows", 42)},
})
```

A zero `Time` omits the timestamp.

### Custom formatters and writers

A custom formatter implements `hqgologgerformatter.Formatter`:

```go
type Formatter interface {
	Format(log hqgologgerformatter.Log) (data []byte, err error)
}
```

Formatters must not mutate the record and must be safe for concurrent use. Implement `formatter.Appender` (`Append(dst []byte, log Log) ([]byte, error)`) to reuse the logger's buffer.

A custom writer implements `hqgologgerwriter.Writer`:

```go
type Writer interface {
	Write(data []byte, l level.Level) error
	Close() error
}
```

Writers must be safe for concurrent use and must not retain `data`. Resource-owning writers should return `writer.ErrClosed` from `Write` after `Close`.

## Contributing

Contributions are welcome and encouraged! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-lib-logger-go/pulls) or report [Issues](https://github.com/hueristiq/hq-lib-logger-go/issues).

A big thank you to all the [contributors](https://github.com/hueristiq/hq-lib-logger-go/graphs/contributors) for your ongoing support!

![contributors](https://contrib.rocks/image?repo=hueristiq/hq-lib-logger-go&max=500)

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-lib-logger-go/blob/main/LICENSE).
