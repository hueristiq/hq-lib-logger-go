// Package level defines the severity levels used by hq-lib-logger-go.
//
// Lower values are more severe: [LevelFatal] (0) is the most severe and
// [LevelTrace] the least.
//
// # Entry points
//
// [Parse] converts a name to a Level, [Level.Valid] reports whether a value is
// defined, and [Level.String] returns the lowercase name.
//
// # Parsing and formatting
//
// Parse trims surrounding whitespace and is case-insensitive. Level implements
// [encoding.TextMarshaler] and [encoding.TextUnmarshaler], so levels round-trip
// through JSON or YAML configuration as their lowercase names.
//
// # Concurrency
//
// Level is an immutable integer value and is safe for concurrent use.
package level
