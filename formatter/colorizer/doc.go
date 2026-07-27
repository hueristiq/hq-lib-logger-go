// Package colorizer provides ANSI color implementations of the
// [github.com/hueristiq/hq-lib-logger-go/formatter.Colorizer] interface.
//
// [NewFatihColorizer] builds a colorizer backed by github.com/fatih/color, and
// [NewAuroraColorizer] one backed by github.com/logrusorgru/aurora. Both map each
// severity level to a distinct color and leave levels.LevelSilent (and any
// invalid level) uncolored. For plain, uncolored output, use
// [github.com/hueristiq/hq-lib-logger-go/formatter.NewNoOpColorizer] instead.
//
// Note that github.com/fatih/color disables color automatically when standard
// output is not a terminal, whereas the aurora-backed colorizer always emits
// escape codes.
package colorizer
