// Package logger writes structured, leveled log records to configurable
// writers.
//
// # Entry points
//
// [New] returns a usable logger at [LevelInfo]. Log records with [Logger.Info],
// [Logger.Warn], [Logger.Error], [Logger.Debug], [Logger.Trace], or [Logger.Log],
// and attach typed [Field] values. [Logger.LogRecord] is the
// advanced path for a full formatter.Log. [Logger.SlogHandler] adapts the logger
// to log/slog.
//
// # Configuration
//
// Pass [OptionFunc] values produced by [WithLevel], [WithFormatter], [WithWriter],
// and [WithCallerCapture] to New. Formatter, writer, and caller capture are fixed
// after construction; only [Logger.SetLevel] changes at runtime. Child views from
// [Logger.Named], [Logger.With], and [Logger.WithPrefix] share one configuration
// core.
//
// # Concurrency
//
// A Logger is safe for concurrent use. Formatters and writers supplied to New
// must provide their own safety.
package logger
