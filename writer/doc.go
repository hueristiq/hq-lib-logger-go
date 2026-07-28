// Package writer delivers formatted log bytes to output destinations.
//
// A [Writer] writes a pre-formatted message together with its severity level and
// can be closed to release resources. [Console] routes output to stdout or stderr
// by level (configurable via [ConsoleWriterConfiguration]), [MultiWriter] fans a
// single message out to several writers at once, and [IOWriter] adapts any
// standard io.Writer (a file, a buffer, a network connection) to the interface.
// Implement [Writer] to add custom destinations such as external logging services.
package writer
