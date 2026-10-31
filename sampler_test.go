//nolint:testpackage // tests exercise unexported behavior.
package logger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSamplerFunc(t *testing.T) {
	t.Parallel()

	assert.True(t, SamplerFunc(nil).Sample(LevelInfo))
	assert.False(t, SamplerFunc(func(Level) bool { return false }).Sample(LevelInfo))
}

func TestBasicSampler(t *testing.T) {
	t.Parallel()

	sampler := NewBasicSampler(2)

	assert.False(t, sampler.Sample(LevelInfo))
	assert.True(t, sampler.Sample(LevelInfo))
	assert.False(t, sampler.Sample(LevelError))
	assert.True(t, sampler.Sample(LevelError))

	assert.True(t, NewBasicSampler(0).Sample(LevelInfo))
	assert.True(t, NewBasicSampler(1).Sample(LevelInfo))

	var nilSampler *BasicSampler

	assert.True(t, nilSampler.Sample(LevelInfo))
}

func TestLevelSampler(t *testing.T) {
	t.Parallel()

	sampler := NewLevelSampler(map[Level]uint64{
		LevelInfo:  2,
		LevelError: 0,
		Level(99):  2,
	})

	assert.False(t, sampler.Sample(LevelInfo))
	assert.True(t, sampler.Sample(LevelInfo))
	assert.True(t, sampler.Sample(LevelError))
	assert.True(t, sampler.Sample(LevelDebug))
	assert.True(t, sampler.Sample(Level(99)))

	var nilSampler *LevelSampler

	assert.True(t, nilSampler.Sample(LevelInfo))
}

func TestRateSampler(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	sampler := NewRateSampler(2, time.Minute)
	sampler.now = func() time.Time { return now }

	assert.True(t, sampler.Sample(LevelInfo))
	assert.True(t, sampler.Sample(LevelInfo))
	assert.False(t, sampler.Sample(LevelInfo))

	now = now.Add(time.Minute)

	assert.True(t, sampler.Sample(LevelInfo))
	assert.True(t, sampler.Sample(LevelInfo))
	assert.False(t, sampler.Sample(LevelInfo))

	assert.True(t, NewRateSampler(0, time.Minute).Sample(LevelInfo))
	assert.True(t, NewRateSampler(1, 0).Sample(LevelInfo))

	var nilSampler *RateSampler

	assert.True(t, nilSampler.Sample(LevelInfo))
}
