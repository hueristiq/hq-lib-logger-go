// Package colorizer provides ANSI color implementations of the
// [github.com/hueristiq/hq-lib-logger-go/formatter.Colorizer] interface.
//
// [NewANSIColorizer] builds a stateless colorizer emitting raw SGR escape
// sequences; it maps each severity level to a distinct high-intensity
// color with bold styling and leaves levels.LevelSilent (and any invalid
// level) uncolored. For plain, uncolored output, use
// [github.com/hueristiq/hq-lib-logger-go/formatter.NewNoOpColorizer]
// instead.
package colorizer
