// Package nilvalue reports whether an interface value holds a nil value of a
// nilable kind.
package nilvalue

import "reflect"

// nilableKinds lists the reflect kinds whose zero value can be nil.
var nilableKinds = map[reflect.Kind]struct{}{
	reflect.Chan:      {},
	reflect.Func:      {},
	reflect.Interface: {},
	reflect.Map:       {},
	reflect.Pointer:   {},
	reflect.Slice:     {},
}

// IsNil reports whether v is nil or holds a typed-nil value of a nilable kind
// (chan, func, interface, map, pointer, or slice). It is used to filter writers
// and render typed-nil errors without calling methods on a nil receiver.
//
// Parameters:
//   - v (any): the value to inspect.
//
// Returns:
//   - ok (bool): true when v is nil or holds a typed-nil nilable value.
func IsNil(v any) (ok bool) {
	if v == nil {
		return true
	}

	value := reflect.ValueOf(v)

	if _, nilable := nilableKinds[value.Kind()]; nilable {
		return value.IsNil()
	}

	return false
}
