package logger

import (
	"sync"
	"sync/atomic"
	"time"
)

// Sampler decides whether a record at a level should be emitted.
//
// Methods:
//   - Sample(level Level) (emit bool): reports whether to emit the record.
type Sampler interface {
	Sample(level Level) (emit bool)
}

// SamplerFunc adapts a function to Sampler. A nil function emits every record.
type SamplerFunc func(level Level) (emit bool)

// Sample calls f(level), or returns true when f is nil.
//
// Parameters:
//   - level (Level): record severity.
//
// Returns:
//   - emit (bool): true when the record should be emitted.
func (f SamplerFunc) Sample(level Level) (emit bool) {
	if f == nil {
		return true
	}

	return f(level)
}

// BasicSampler emits every Nth record across all levels.
//
// Fields:
//   - n (uint64): emit interval.
//   - counter (atomic.Uint64): number of sampled records.
type BasicSampler struct {
	n       uint64
	counter atomic.Uint64
}

// NewBasicSampler returns a sampler that emits every nth record. Zero or one
// emits every record.
//
// Parameters:
//   - n (uint64): emit interval.
//
// Returns:
//   - sampler (*BasicSampler): the sampler.
func NewBasicSampler(n uint64) (sampler *BasicSampler) {
	return &BasicSampler{n: n}
}

// Sample reports whether the record should be emitted.
//
// Parameters:
//   - level (Level): record severity; ignored.
//
// Returns:
//   - emit (bool): true for every nth record.
func (s *BasicSampler) Sample(_ Level) (emit bool) {
	if s == nil || s.n <= 1 {
		return true
	}

	return s.counter.Add(1)%s.n == 0
}

// LevelSampler emits every Nth record independently per level.
//
// Fields:
//   - divisors ([6]uint64): per-level emit intervals.
//   - counters ([6]atomic.Uint64): per-level sampled counts.
type LevelSampler struct {
	divisors [6]uint64
	counters [6]atomic.Uint64
}

// NewLevelSampler returns a per-level sampler. Missing, zero, or one divisors
// emit every record for that level.
//
// Parameters:
//   - divisors (map[Level]uint64): emit interval per level.
//
// Returns:
//   - sampler (*LevelSampler): the sampler.
func NewLevelSampler(divisors map[Level]uint64) (sampler *LevelSampler) {
	sampler = &LevelSampler{}

	for level, divisor := range divisors {
		if level.Valid() {
			sampler.divisors[int(level)] = divisor
		}
	}

	return sampler
}

// Sample reports whether the record should be emitted.
//
// Parameters:
//   - level (Level): record severity.
//
// Returns:
//   - emit (bool): true for every nth record at level.
func (s *LevelSampler) Sample(level Level) (emit bool) {
	if s == nil || !level.Valid() {
		return true
	}

	divisor := s.divisors[int(level)]
	if divisor <= 1 {
		return true
	}

	return s.counters[int(level)].Add(1)%divisor == 0
}

// RateSampler emits at most limit records per fixed interval.
//
// Fields:
//   - mutex (sync.Mutex): guards the window state.
//   - limit (int): maximum records per interval.
//   - interval (time.Duration): window length.
//   - windowStart (time.Time): start of the current window.
//   - count (int): records emitted in the current window.
//   - now (func() time.Time): clock override for tests.
type RateSampler struct {
	mutex       sync.Mutex
	limit       int
	interval    time.Duration
	windowStart time.Time
	count       int
	now         func() time.Time
}

// NewRateSampler returns a fixed-window rate sampler. A non-positive limit or
// interval emits every record.
//
// Parameters:
//   - limit (int): maximum records per interval.
//   - interval (time.Duration): window length.
//
// Returns:
//   - sampler (*RateSampler): the sampler.
func NewRateSampler(limit int, interval time.Duration) (sampler *RateSampler) {
	return &RateSampler{
		limit:    limit,
		interval: interval,
		now:      time.Now,
	}
}

// Sample reports whether the record should be emitted.
//
// Parameters:
//   - level (Level): record severity; ignored.
//
// Returns:
//   - emit (bool): true while the current window has capacity.
func (s *RateSampler) Sample(_ Level) (emit bool) {
	if s == nil || s.limit <= 0 || s.interval <= 0 {
		return true
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now
	if s.now != nil {
		now = s.now
	}

	current := now()
	if s.windowStart.IsZero() || current.Sub(s.windowStart) >= s.interval {
		s.windowStart = current
		s.count = 0
	}

	if s.count >= s.limit {
		return false
	}

	s.count++

	return true
}

var (
	_ Sampler = (*BasicSampler)(nil)
	_ Sampler = (*LevelSampler)(nil)
	_ Sampler = (*RateSampler)(nil)
	_ Sampler = SamplerFunc(nil)
)
