package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/hueristiq/hq-lib-logger-go/internal/nilvalue"
)

type fieldKind uint8

const (
	kindString fieldKind = iota
	kindInt
	kindUint
	kindFloat
	kindBool
	kindDuration
	kindTime
	kindError
	kindAny
)

// Field is an opaque typed key-value pair attached to a log record.
//
// Fields are immutable. Construct them with [String], [Int], [Int64], [Uint],
// [Uint64], [Float64], [Bool], [Duration], [Time], [Err], or [Any].
// Duplicate keys are preserved in insertion order.
//
// Fields:
//   - key (string): field name.
//   - kind (fieldKind): internal value representation.
//   - str (string): string value for KindString.
//   - integer (int64): signed value for KindInt, KindBool, KindDuration, and KindTime.
//   - uinteger (uint64): unsigned value for KindUint.
//   - float (float64): floating-point value for KindFloat.
//   - any (any): arbitrary value for KindError and KindAny.
type Field struct {
	key      string
	kind     fieldKind
	str      string
	integer  int64
	uinteger uint64
	float    float64
	any      any
}

// String returns a string field.
//
// Parameters:
//   - key (string): field name.
//   - value (string): field value.
//
// Returns:
//   - field (Field): a string field.
func String(key, value string) (field Field) {
	return Field{key: key, kind: kindString, str: value}
}

// Int returns an integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (int): field value.
//
// Returns:
//   - field (Field): an integer field.
func Int(key string, value int) (field Field) {
	return Int64(key, int64(value))
}

// Int64 returns a signed 64-bit integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (int64): field value.
//
// Returns:
//   - field (Field): an integer field.
func Int64(key string, value int64) (field Field) {
	return Field{key: key, kind: kindInt, integer: value}
}

// Uint returns an unsigned integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (uint): field value.
//
// Returns:
//   - field (Field): an unsigned integer field.
func Uint(key string, value uint) (field Field) {
	return Uint64(key, uint64(value))
}

// Uint64 returns an unsigned 64-bit integer field.
//
// Parameters:
//   - key (string): field name.
//   - value (uint64): field value.
//
// Returns:
//   - field (Field): an unsigned integer field.
func Uint64(key string, value uint64) (field Field) {
	return Field{key: key, kind: kindUint, uinteger: value}
}

// Float64 returns a floating-point field.
//
// Parameters:
//   - key (string): field name.
//   - value (float64): field value.
//
// Returns:
//   - field (Field): a floating-point field.
func Float64(key string, value float64) (field Field) {
	return Field{key: key, kind: kindFloat, float: value}
}

// Bool returns a boolean field.
//
// Parameters:
//   - key (string): field name.
//   - value (bool): field value.
//
// Returns:
//   - field (Field): a boolean field.
func Bool(key string, value bool) (field Field) {
	f := Field{key: key, kind: kindBool}
	if value {
		f.integer = 1
	}

	return f
}

// Duration returns a duration field.
//
// Parameters:
//   - key (string): field name.
//   - value (time.Duration): field value.
//
// Returns:
//   - field (Field): a duration field.
func Duration(key string, value time.Duration) (field Field) {
	return Field{key: key, kind: kindDuration, integer: int64(value)}
}

// Time returns a time field. The value is stored as Unix nanoseconds.
//
// Parameters:
//   - key (string): field name.
//   - value (time.Time): field value.
//
// Returns:
//   - field (Field): a time field.
func Time(key string, value time.Time) (field Field) {
	return Field{key: key, kind: kindTime, integer: value.UnixNano()}
}

// Err returns an error field.
//
// Parameters:
//   - key (string): field name.
//   - value (error): field value.
//
// Returns:
//   - field (Field): an error field.
func Err(key string, value error) (field Field) {
	return Field{key: key, kind: kindError, any: value}
}

// Any returns a field backed by an arbitrary value.
//
// Parameters:
//   - key (string): field name.
//   - value (any): field value.
//
// Returns:
//   - field (Field): an arbitrary-value field.
func Any(key string, value any) (field Field) {
	return Field{key: key, kind: kindAny, any: value}
}

// Key returns the field key.
//
// Returns:
//   - key (string): the field name.
//
//nolint:gocritic // Field is intentionally a value type.
func (f Field) Key() (key string) {
	return f.key
}

// Value returns the field's value as an any.
//
// Returns:
//   - value (any): the field value.
//
//nolint:gocritic // Field is intentionally a value type.
func (f Field) Value() (value any) {
	switch f.kind {
	case kindString:
		return f.str
	case kindInt:
		return f.integer
	case kindUint:
		return f.uinteger
	case kindFloat:
		return f.float
	case kindBool:
		return f.integer != 0
	case kindDuration:
		return time.Duration(f.integer)
	case kindTime:
		return time.Unix(0, f.integer).UTC()
	case kindError, kindAny:
		return f.any
	default:
		// Unreachable: every kind is listed above. Kept defensively.
		return nil
	}
}

// AppendText appends the field value's text representation.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//
// Returns:
//   - out ([]byte): dst with the value appended.
//
//nolint:gocritic // Field is intentionally a value type.
func (f Field) AppendText(dst []byte) (out []byte) {
	switch f.kind {
	case kindString:
		return append(dst, f.str...)
	case kindInt:
		return strconv.AppendInt(dst, f.integer, 10)
	case kindUint:
		return strconv.AppendUint(dst, f.uinteger, 10)
	case kindFloat:
		return strconv.AppendFloat(dst, f.float, 'g', -1, 64)
	case kindBool:
		return strconv.AppendBool(dst, f.integer != 0)
	case kindDuration:
		return append(dst, time.Duration(f.integer).String()...)
	case kindTime:
		return time.Unix(0, f.integer).UTC().AppendFormat(dst, time.RFC3339Nano)
	case kindError:
		return append(dst, nilSafeErrorText(f.any)...)
	case kindAny:
		return fmt.Append(dst, f.any)
	default:
		// Unreachable: every kind is listed above. Kept defensively.
		return dst
	}
}

// AppendJSON appends the field value as JSON.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//
// Returns:
//   - out ([]byte): dst with the JSON value appended.
//   - err (error): non-nil when an arbitrary value cannot be marshaled.
//
//nolint:gocritic // Field is intentionally a value type.
func (f Field) AppendJSON(dst []byte) (out []byte, err error) {
	switch f.kind {
	case kindString:
		return appendJSONString(dst, f.str), nil
	case kindInt:
		return strconv.AppendInt(dst, f.integer, 10), nil
	case kindUint:
		return strconv.AppendUint(dst, f.uinteger, 10), nil
	case kindFloat:
		if f.float != f.float || f.float > 1.7976931348623157e+308 || f.float < -1.7976931348623157e+308 {
			return append(dst, "null"...), nil
		}

		return strconv.AppendFloat(dst, f.float, 'g', -1, 64), nil
	case kindBool:
		return strconv.AppendBool(dst, f.integer != 0), nil
	case kindDuration:
		return appendJSONString(dst, time.Duration(f.integer).String()), nil
	case kindTime:
		return appendJSONString(dst, time.Unix(0, f.integer).UTC().Format(time.RFC3339Nano)), nil
	case kindError:
		return appendJSONString(dst, nilSafeErrorText(f.any)), nil
	case kindAny:
		data, marshalErr := json.Marshal(f.any)
		if marshalErr != nil {
			return dst, marshalErr
		}

		return append(dst, data...), nil
	default:
		// Unreachable: every kind is listed above. Kept defensively.
		return append(dst, "null"...), nil
	}
}

// isNilAny reports whether f holds a nil value of a nilable kind.
//
// isNilAny reports whether f holds a nil value of a nilable kind.
//
// Returns:
//   - nil (bool): true when the field holds a nil value.
//
//nolint:gocritic // Field is intentionally a value type.
func (f Field) isNilAny() bool {
	return (f.kind == kindAny || f.kind == kindError) && nilvalue.IsNil(f.any)
}

// nilSafeErrorText renders a value without panicking on typed-nil errors.
//
// Parameters:
//   - value (any): value to render.
//
// Returns:
//   - text (string): the rendered value, or "<nil>" for nil values.
func nilSafeErrorText(value any) string {
	if nilvalue.IsNil(value) {
		return "<nil>"
	}

	if err, ok := value.(error); ok {
		return err.Error()
	}

	return fmt.Sprint(value)
}

// appendJSONString appends s as a valid JSON string without reflection.
//
// Parameters:
//   - dst ([]byte): destination buffer.
//   - s (string): string to encode.
//
// Returns:
//   - out ([]byte): dst with the JSON-encoded string appended.
func appendJSONString(dst []byte, s string) []byte {
	dst = append(dst, '"')

	for i := 0; i < len(s); {
		c := s[i]

		switch c {
		case '"', '\\':
			dst = append(dst, '\\', c)
			i++

			continue
		case '\n':
			dst = append(dst, '\\', 'n')
			i++

			continue
		case '\r':
			dst = append(dst, '\\', 'r')
			i++

			continue
		case '\t':
			dst = append(dst, '\\', 't')
			i++

			continue
		case '\b':
			dst = append(dst, '\\', 'b')
			i++

			continue
		case '\f':
			dst = append(dst, '\\', 'f')
			i++

			continue
		}

		if c < 0x20 {
			dst = append(dst, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xF])
			i++

			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			dst = append(dst, `\ufffd`...)
			i++

			continue
		}

		dst = append(dst, s[i:i+size]...)
		i += size
	}

	return append(dst, '"')
}

const hex = "0123456789abcdef"

var fieldIndicesPool = sync.Pool{
	New: func() any {
		indices := make([]int, 0, 16)

		return &indices
	},
}

// acquireFieldIndices returns a pooled index slice of length n for
// zero-allocation field sorting.
//
// Parameters:
//   - n (int): number of indices needed.
//
// Returns:
//   - indices (*[]int): the pooled slice, grown when its capacity is
//     insufficient.
func acquireFieldIndices(n int) (indices *[]int) {
	indices = fieldIndicesPool.Get().(*[]int)

	if cap(*indices) < n {
		*indices = make([]int, n)
	} else {
		*indices = (*indices)[:n]
	}

	return indices
}

// releaseFieldIndices returns an index slice to the pool.
//
// Parameters:
//   - indices (*[]int): slice to return.
func releaseFieldIndices(indices *[]int) {
	fieldIndicesPool.Put(indices)
}
