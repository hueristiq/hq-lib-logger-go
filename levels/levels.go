package levels

import (
	"encoding"
	"errors"
	"fmt"
	"strings"
)

// Level is a log severity. Lower values are more severe:
// [LevelFatal] is 0 and [LevelTrace] is the least severe.
type Level int

const (
	// LevelFatal is the most severe level; logging at it exits the process.
	LevelFatal Level = iota
	// LevelError reports errors that do not halt the program.
	LevelError
	// LevelWarn reports conditions that may need attention.
	LevelWarn
	// LevelInfo reports normal operation.
	LevelInfo
	// LevelDebug reports diagnostic detail.
	LevelDebug
	// LevelTrace reports the most verbose detail.
	LevelTrace
)

// ErrUnknownLevel is returned when a level name or value is not recognized.
var ErrUnknownLevel = errors.New("unknown level")

// levelNames maps valid levels to their lowercase names.
var levelNames = [...]string{"fatal", "error", "warn", "info", "debug", "trace"}

// String returns the lowercase name of l, or "unknown" when l is invalid.
//
// Returns:
//   - name (string): the lowercase level name or "unknown".
func (l Level) String() (name string) {
	if !l.Valid() {
		return "unknown"
	}

	return levelNames[int(l)]
}

// Valid reports whether l is one of the defined levels.
//
// Returns:
//   - valid (bool): true when l is a defined level.
func (l Level) Valid() (valid bool) {
	return l >= LevelFatal && l <= LevelTrace
}

// Parse parses a level name. Matching is case-insensitive and surrounding
// whitespace is ignored.
//
// Parameters:
//   - name (string): level name such as "info" or "trace".
//
// Returns:
//   - level (Level): the parsed level, or LevelFatal on error.
//   - err (error): non-nil wrapping [ErrUnknownLevel] for an unknown name.
func Parse(name string) (level Level, err error) {
	normalized := strings.ToLower(strings.TrimSpace(name))

	for i, levelName := range levelNames {
		if levelName == normalized {
			return Level(i), nil
		}
	}

	return LevelFatal, fmt.Errorf("%w (%q)", ErrUnknownLevel, name)
}

// MarshalText implements encoding.TextMarshaler. It returns
// [ErrUnknownLevel] for an invalid level.
//
// Returns:
//   - text ([]byte): the lowercase level name.
//   - err (error): non-nil wrapping [ErrUnknownLevel] for an invalid level.
func (l Level) MarshalText() (text []byte, err error) {
	if !l.Valid() {
		return nil, fmt.Errorf("%w (%d)", ErrUnknownLevel, int(l))
	}

	return []byte(l.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler using [Parse].
//
// Parameters:
//   - text ([]byte): level name to parse.
//
// Returns:
//   - err (error): non-nil wrapping [ErrUnknownLevel] for an unknown name.
func (l *Level) UnmarshalText(text []byte) (err error) {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}

	*l = parsed

	return nil
}

var (
	_ encoding.TextMarshaler   = Level(0)
	_ encoding.TextUnmarshaler = (*Level)(nil)
)
