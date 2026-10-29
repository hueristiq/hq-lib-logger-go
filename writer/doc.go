// Package writer delivers formatted log bytes to output destinations.
//
// # Entry points
//
// A [Writer] writes a preformatted message with its severity and can be closed.
// [NewStream] adapts any io.Writer, [NewMulti] fans out to several writers,
// [NewSyslog] writes to Unix syslog, and [NewEventLog] writes to the Windows
// event log.
//
// # Configuration
//
// [SyslogConfig] and [EventLogConfig] configure the system sinks. [NewStream]
// and [NewMulti] need no configuration. Platform sinks return
// [ErrUnsupported] where unavailable.
//
// # Concurrency
//
// Writers are called concurrently and must be safe for concurrent use. Write
// must not retain the data slice after it returns.
package writer
