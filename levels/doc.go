// Package levels defines the severity levels used throughout hq-lib-logger-go.
//
// A [Level] is an integer in which lower values are more severe: [LevelFatal]
// (0) is the most severe and [LevelDebug] (5) the least. [LevelSilent] (1) is
// special — as a logger threshold it suppresses every level except LevelFatal
// and itself, while as a message level it marks user-facing "print" output.
// Level implements [encoding.TextMarshaler] and [encoding.TextUnmarshaler],
// so levels round-trip through JSON or YAML configuration as their lowercase
// names ("fatal", "info", and so on).
package levels
