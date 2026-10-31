package nilvalue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type nilMap map[string]int

func TestIsNil(t *testing.T) {
	t.Parallel()

	var (
		nilMapValue    nilMap
		nilSlice       []int
		nilFunc        func()
		nilChan        chan int
		nilPointer     *int
		nonNilPointer  = new(int)
		nonNilMap      = nilMap{}
		nonNilSlice    = []int{}
		nonNilFunc     = func() {}
		nonNilChan     = make(chan int)
		nonNilableType = struct{}{}
	)

	assert.True(t, IsNil(nil))
	assert.True(t, IsNil(nilMapValue))
	assert.True(t, IsNil(nilSlice))
	assert.True(t, IsNil(nilFunc))
	assert.True(t, IsNil(nilChan))
	assert.True(t, IsNil(nilPointer))

	assert.False(t, IsNil(nonNilPointer))
	assert.False(t, IsNil(nonNilMap))
	assert.False(t, IsNil(nonNilSlice))
	assert.False(t, IsNil(nonNilFunc))
	assert.False(t, IsNil(nonNilChan))
	assert.False(t, IsNil(nonNilableType))
}
