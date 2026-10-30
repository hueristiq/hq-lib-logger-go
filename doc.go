// Package logger provides structured, leveled logging.
//
// # Entry points
//
// [New] returns a usable logger at [LevelInfo]. Log records with [Logger.Info],
// [Logger.Warn], [Logger.Error], [Logger.Debug], [Logger.Trace], [Logger.Fatal],
// or [Logger.Log], and attach typed [Field] values. [Logger.LogRecord] is the
// advanced path for a full formatter.Log. [Logger.SlogHandler] adapts the logger
// to log/slog.
//
// # Configuration
//
// Pass [OptionFunc] values produced by [WithLevel], [WithFormatter], [WithWriter],
// [WithSampler], [WithHooks], and [WithCallerCapture] to New. Formatter, writer,
// sampler, hooks, and caller capture are fixed after construction; only
// [Logger.SetLevel] changes at runtime. Child views from [Logger.Named],
// [Logger.With], and [Logger.WithPrefix] share one configuration core.
//
// # Concurrency
//
// A Logger is safe for concurrent use. Formatters, writers, hooks, and samplers
// supplied to New must provide their own safety. Hooks and samplers run while
// the logger holds its read lock and must not call logger configuration methods.
package logger
