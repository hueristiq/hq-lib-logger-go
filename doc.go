// Package logger provides flexible, extensible structured logging.
//
// A [Logger] processes every message through a three-stage pipeline: it filters
// events against a severity threshold, hands the survivors to a
// [github.com/hueristiq/hq-lib-logger-go/formatter.Formatter] that renders them
// to bytes, and passes those bytes to a
// [github.com/hueristiq/hq-lib-logger-go/writer.Writer] for delivery. Each stage
// is an interface, so formatters, writers, and label colorizers can be swapped
// independently.
//
// # Quick start
//
// The package-level functions ([Info], [Warn], [Error], [Debug], [Print],
// [Fatal]) log through [DefaultLogger], which is preconfigured with a debug
// threshold, a console formatter (RFC3339 timestamps and labels), and a console
// writer that sends [Print] output to stdout and everything else to stderr:
//
//	logger.Info("listening", logger.WithString("addr", ":8080"))
//	logger.Error("query failed", logger.WithError(err))
//
// For full control, construct a [Logger] with [NewLogger] and set its level,
// formatter, and writer explicitly. A freshly constructed Logger has no formatter
// or writer and silently drops events until both are configured.
//
// # Severity levels
//
// Levels are defined in the
// [github.com/hueristiq/hq-lib-logger-go/levels] package and ordered so that
// lower values are more severe (LevelFatal = 0 through LevelDebug = 5). A logger
// emits an event only when its level is at least as severe as the threshold, that
// is, when event level <= configured level. [Logger.Fatal] writes its message and
// then calls os.Exit(1), so reserve it for unrecoverable conditions.
//
// # Options
//
// Each logging call accepts zero or more [OptionFunc] values that attach metadata
// or adjust rendering: [WithString], [WithValue], and [WithError] add metadata,
// while [WithLabel], [WithoutLabel], and [WithoutTimestamp] control the label and
// timestamp of a single event.
//
// # Concurrency
//
// A [Logger] is safe for concurrent use by multiple goroutines: its configuration
// is guarded by a sync.RWMutex, and the bundled console writer serializes its
// writes with a mutex. Formatters and writers supplied by callers are expected to
// provide their own safety.
package logger
