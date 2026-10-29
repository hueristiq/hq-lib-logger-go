# hq-lib-logger-go

![made with go](https://img.shields.io/badge/made%20with-Go-1E90FF.svg) [![go reference](https://pkg.go.dev/badge/github.com/hueristiq/hq-lib-logger-go.svg)](https://pkg.go.dev/github.com/hueristiq/hq-lib-logger-go) [![license](https://img.shields.io/badge/license-MIT-gray.svg?color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/blob/main/LICENSE) ![maintenance](https://img.shields.io/badge/maintained%3F-yes-1E90FF.svg) [![open issues](https://img.shields.io/github/issues-raw/hueristiq/hq-lib-logger-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/issues?q=is:issue+is:open) [![closed issues](https://img.shields.io/github/issues-closed-raw/hueristiq/hq-lib-logger-go.svg?style=flat&color=1E90FF)](https://github.com/hueristiq/hq-lib-logger-go/issues?q=is:issue+is:closed) [![contribution](https://img.shields.io/badge/contributions-welcome-1E90FF.svg)](https://github.com/hueristiq/hq-lib-logger-go/blob/main/CONTRIBUTING.md)

`hq-lib-logger-go` is a [Go](https://golang.org/) package for structured logging.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
	- [Quick Start](#quick-start)
	- [Log Levels](#log-levels)
	- [Structured Fields](#structured-fields)
	- [Child Loggers](#child-loggers)
	- [Formatters](#formatters)
	- [Writers](#writers)
	- [Sampling and Hooks](#sampling-and-hooks)
	- [log/slog Integration](#logslog-integration)
	- [Context Loggers](#context-loggers)
	- [Building a Logger](#building-a-logger)
	- [Custom Formatters and Writers](#custom-formatters-and-writers)
- [Contributing](#contributing)
- [Licensing](#licensing)

## Features

- **Six Severity Levels:** `Fatal`, `Error`, `Warn`, `Info`, `Debug`, and `Trace`, behind a configurable threshold.
- **Typed Fields:** `String`, `Int`, `Int64`, `Uint`, `Uint64`, `Float64`, `Bool`, `Duration`, `Time`, `Err`, and `Any`.
- **Child Views:** `Named`, `With`, and `WithPrefix` share one configuration core.
- **Pluggable Output:** console, JSON, and logfmt formatters; stream, fan-out, syslog, and Windows event log writers.
- **Sampling and Hooks:** basic, per-level, and rate samplers; pre-format hooks.
- **Ecosystem Integration:** a `log/slog` handler and context helpers.
- **Opaque Internals:** sampler state and fields are immutable from outside the package.
- **Standard-Library Runtime:** the optional Windows event log writer uses `golang.org/x/sys` on Windows.

## Installation

```bash
go get github.com/hueristiq/hq-lib-logger-go@latest
```

Requires Go 1.23 or later.

## Usage

The examples import the package under the `hqgologger` alias; subpackages use `hqgologgerformatter`, `hqgologgerlevels`, and `hqgologgerwriter`:

```go
import (
	hqgologger "github.com/hueristiq/hq-lib-logger-go"
	hqgologgerformatter "github.com/hueristiq/hq-lib-logger-go/formatter"
	hqgologgerlevels "github.com/hueristiq/hq-lib-logger-go/levels"
	hqgologgerwriter "github.com/hueristiq/hq-lib-logger-go/writer"
)
```

### Quick Start

`New()` is usable immediately. It uses a console formatter, writes to stderr, and starts at `LevelInfo`:

```go
log := hqgologger.New()

log.Info("application started", hqgologger.String("service", "api"))
log.Error("request failed",
	hqgologger.Err("error", err),
	hqgologger.Int("status", 500),
)
```

Configure the logger at construction when the defaults do not fit:

```go
log := hqgologger.New(
	hqgologger.WithLevel(hqgologgerlevels.LevelDebug),
	hqgologger.WithFormatter(hqgologgerformatter.NewJSON(nil)),
	hqgologger.WithWriter(hqgologgerwriter.NewStream(os.Stdout)),
)
```

### Log Levels

| Level | Value | Meaning |
| --- | --- | --- |
| `LevelFatal` | 0 | Unrecoverable failure; `Fatal` exits the process. |
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

`Fatal` always writes and calls `os.Exit(1)`, bypassing sampling.

### Structured Fields

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

### Child Loggers

`With` attaches fields, `Named` adds a dotted logger name, and `WithPrefix` prepends raw text to messages:

```go
api := log.Named("api").With(hqgologger.String("service", "api"))

api.Info("started")
api.WithPrefix("[request] ").Info("finished")
```

Child views share the logger's level, formatter, writer, sampler, and hooks. Fields, name, and prefix belong to the view.

### Formatters

```go
hqgologgerformatter.NewConsole(nil) // timestamp, label, sorted fields
hqgologgerformatter.NewJSON(nil)    // one JSON object per record
hqgologgerformatter.NewLogfmt(nil)  // compact key=value
```

Each constructor accepts a nil config for package defaults. A non-nil zero config is explicit: timestamps and labels are off, optional keys are omitted. For example, this JSON formatter emits no timestamp and nests fields:

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

### Sampling and Hooks

```go
log := hqgologger.New(
	hqgologger.WithSampler(hqgologger.NewRateSampler(100, time.Second)),
	hqgologger.WithHooks(hqgologger.HookFunc(func(record *hqgologgerformatter.Log) error {
		return nil
	})),
)
```

`NewBasicSampler(n)` emits every Nth event, `NewLevelSampler(map[Level]uint64)` samples per level, and `NewRateSampler(limit, interval)` enforces a fixed window. `Fatal` bypasses sampling. Hooks run before formatting; hook errors are ignored.

### log/slog Integration

```go
slog.SetDefault(slog.New(log.SlogHandler()))

slog.Info("hello", "request_id", "42")
```

The handler maps slog levels to logger levels, converts attributes and groups to typed fields, and never exits for high slog levels.

### Global Logger

The package intentionally has no mutable package-level logger. For `slog.Info`-style convenience, install a default handler once at startup:

```go
slog.SetDefault(log.SlogHandler())

slog.Info("hello", "request_id", "42") // package-level, goes through log
```

### Context Loggers

```go
ctx = hqgologger.NewContext(ctx, log)

log, ok := hqgologger.FromContext(ctx)
if !ok {
	// no logger stored
}
```

`NewContext` with a nil logger returns the context unchanged.

### Building a Logger

```go
log := hqgologger.New(
	hqgologger.WithLevel(hqgologgerlevels.LevelInfo),
	hqgologger.WithFormatter(hqgologgerformatter.NewJSON(nil)),
	hqgologger.WithWriter(hqgologgerwriter.NewStream(os.Stdout)),
	hqgologger.WithSampler(hqgologger.NewBasicSampler(1)),
	hqgologger.WithHooks(hqgologger.HookFunc(hostnameHook)),
	hqgologger.WithCallerCapture(true),
)
```

Options are applied in order. Nil formatters, writers, hooks, and samplers are ignored; an invalid level is ignored and the default `LevelInfo` remains. `SetLevel` is the only runtime configuration mutator; create a new logger to change the formatter, writer, sampler, hooks, or caller capture.

`LogRecord` is the advanced path for a full `hqgologgerformatter.Log`, including a custom time, caller, and name:

```go
log.LogRecord(hqgologgerformatter.Log{
	Level:   hqgologger.LevelInfo,
	Message: "import finished",
	Fields:  []hqgologger.Field{hqgologger.Int("rows", 42)},
})
```

A zero `Time` omits the timestamp.

### Custom Formatters and Writers

A custom formatter implements `hqgologgerformatter.Formatter`:

```go
type Formatter interface {
	Format(log *formatter.Log) (data []byte, err error)
}
```

Formatters must reject a nil `*Log` with `formatter.ErrNilLog`, must not mutate the record, and must be safe for concurrent use. Implement `formatter.Appender` (`Append(dst []byte, log Log) ([]byte, error)`) to reuse the logger's buffer.

A custom writer implements `hqgologgerwriter.Writer`:

```go
type Writer interface {
	Write(data []byte, level levels.Level) error
	Close() error
}
```

Writers must be safe for concurrent use and must not retain `data`. Resource-owning writers should return `writer.ErrClosed` from `Write` after `Close`.

## Contributing

Contributions are welcome and encouraged! Feel free to submit [Pull Requests](https://github.com/hueristiq/hq-lib-logger-go/pulls) or report [Issues](https://github.com/hueristiq/hq-lib-logger-go/issues). For more details, check out the [contribution guidelines](https://github.com/hueristiq/hq-lib-logger-go/blob/main/CONTRIBUTING.md).

A big thank you to all the [contributors](https://github.com/hueristiq/hq-lib-logger-go/graphs/contributors) for your ongoing support!

![contributors](https://contrib.rocks/image?repo=hueristiq/hq-lib-logger-go&max=500)

## Licensing

This package is licensed under the [MIT license](https://opensource.org/license/mit). You are free to use, modify, and distribute it, as long as you follow the terms of the license. You can find the full license text in the repository - [Full MIT license text](https://github.com/hueristiq/hq-lib-logger-go/blob/main/LICENSE).
