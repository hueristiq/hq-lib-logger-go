package writer

import "errors"

var (
	// ErrUnsupported is returned by platform-specific writers on platforms that
	// do not provide the underlying sink.
	ErrUnsupported = errors.New("writer: unsupported on this platform")
	// ErrClosed is returned when writing to a closed system writer.
	ErrClosed = errors.New("writer: closed")
)
