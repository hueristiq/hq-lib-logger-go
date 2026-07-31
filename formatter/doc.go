// Package formatter renders log events into bytes for output.
//
// A [Formatter] converts a [Log] into a byte slice. [Console] is the bundled
// implementation, producing human-readable "[timestamp] [label] message metadata"
// lines with sorted metadata and an optional trailing error block. The metadata
// keys [LabelKey] and [ErrorKey] are reserved: they drive the bracketed label and
// the error block instead of rendering as key=value pairs, and [Console]
// substitutes a level-based default label when a log carries none. Label
// colorization is delegated to a [Colorizer]: the default [NoOpColorizer] leaves
// text unchanged, while the
// [github.com/hueristiq/hq-lib-logger-go/formatter/colorizer] package supplies
// ANSI-coloring implementations. Implement [Formatter] to emit other layouts such
// as JSON or Logfmt.
package formatter
