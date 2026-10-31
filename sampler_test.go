package logger_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	hqgologger "github.com/hueristiq/hq-lib-logger-go"
	hqgologgerlevel "github.com/hueristiq/hq-lib-logger-go/level"
)

func TestBasicSampler(t *testing.T) {
	t.Parallel()

	sampler := hqgologger.NewBasicSampler(2)

	assert.False(t, sampler.Sample(hqgologgerlevel.LevelInfo))
	assert.True(t, sampler.Sample(hqgologgerlevel.LevelInfo))
	assert.False(t, sampler.Sample(hqgologgerlevel.LevelInfo))

	assert.True(t, hqgologger.NewBasicSampler(0).Sample(hqgologgerlevel.LevelInfo))
	assert.True(t, hqgologger.NewBasicSampler(1).Sample(hqgologgerlevel.LevelInfo))
}

func TestLevelSampler(t *testing.T) {
	t.Parallel()

	sampler := hqgologger.NewLevelSampler(map[hqgologgerlevel.Level]uint64{
		hqgologgerlevel.LevelInfo: 2,
	})

	assert.False(t, sampler.Sample(hqgologgerlevel.LevelInfo))
	assert.True(t, sampler.Sample(hqgologgerlevel.LevelInfo))
	assert.True(t, sampler.Sample(hqgologgerlevel.LevelError))
	assert.True(t, sampler.Sample(hqgologgerlevel.Level(99)))
}

func TestRateSampler(t *testing.T) {
	t.Parallel()

	sampler := hqgologger.NewRateSampler(2, time.Hour)

	assert.True(t, sampler.Sample(hqgologgerlevel.LevelInfo))
	assert.True(t, sampler.Sample(hqgologgerlevel.LevelInfo))
	assert.False(t, sampler.Sample(hqgologgerlevel.LevelInfo))

	unbounded := hqgologger.NewRateSampler(0, 0)

	assert.True(t, unbounded.Sample(hqgologgerlevel.LevelInfo))
}

func TestSamplerFunc(t *testing.T) {
	t.Parallel()

	var sampler hqgologger.SamplerFunc = func(level hqgologgerlevel.Level) bool {
		return level == hqgologgerlevel.LevelError
	}

	assert.True(t, sampler.Sample(hqgologgerlevel.LevelError))
	assert.False(t, sampler.Sample(hqgologgerlevel.LevelInfo))

	var nilSampler hqgologger.SamplerFunc

	assert.True(t, nilSampler.Sample(hqgologgerlevel.LevelInfo))
}
