package analyzer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_NoLimit(t *testing.T) {
	rl := NewRateLimiter(0, 0)
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 5; i++ {
		err := rl.Wait(ctx)
		require.NoError(t, err)
	}
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 10*time.Millisecond, "should not delay when no limit is set")
}

func TestRateLimiter_WithDelay(t *testing.T) {
	delay := 50 * time.Millisecond
	rl := NewRateLimiter(delay, 0)
	ctx := context.Background()

	var timestamps []time.Time
	for i := 0; i < 3; i++ {
		err := rl.Wait(ctx)
		require.NoError(t, err)
		timestamps = append(timestamps, time.Now())
	}

	for i := 1; i < len(timestamps); i++ {
		interval := timestamps[i].Sub(timestamps[i-1])
		assert.GreaterOrEqual(t, interval, delay-5*time.Millisecond,
			"interval between requests should be at least the delay")
	}
}

func TestRateLimiter_WithRPS(t *testing.T) {
	rps := 10.0
	expectedInterval := time.Duration(float64(time.Second) / rps)
	rl := NewRateLimiter(0, rps)
	ctx := context.Background()

	var timestamps []time.Time
	for i := 0; i < 3; i++ {
		err := rl.Wait(ctx)
		require.NoError(t, err)
		timestamps = append(timestamps, time.Now())
	}

	for i := 1; i < len(timestamps); i++ {
		interval := timestamps[i].Sub(timestamps[i-1])
		assert.GreaterOrEqual(t, interval, expectedInterval-5*time.Millisecond,
			"interval between requests should match RPS")
	}
}

func TestRateLimiter_RPSOverridesDelay(t *testing.T) {
	delay := 500 * time.Millisecond
	rps := 20.0
	expectedInterval := time.Duration(float64(time.Second) / rps)

	rl := NewRateLimiter(delay, rps)

	assert.Equal(t, expectedInterval, rl.Interval(),
		"RPS should override delay when both are specified")
}

func TestRateLimiter_ContextCancel(t *testing.T) {
	delay := 1 * time.Second
	rl := NewRateLimiter(delay, 0)

	ctx, cancel := context.WithCancel(context.Background())

	err := rl.Wait(ctx)
	require.NoError(t, err)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err = rl.Wait(ctx)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, 100*time.Millisecond,
		"should return immediately when context is canceled")
}

func TestRateLimiter_Interval(t *testing.T) {
	tests := []struct {
		name     string
		delay    time.Duration
		rps      float64
		expected time.Duration
	}{
		{
			name:     "no limit",
			delay:    0,
			rps:      0,
			expected: 0,
		},
		{
			name:     "delay only",
			delay:    200 * time.Millisecond,
			rps:      0,
			expected: 200 * time.Millisecond,
		},
		{
			name:     "rps only",
			delay:    0,
			rps:      5,
			expected: 200 * time.Millisecond,
		},
		{
			name:     "rps overrides delay",
			delay:    1 * time.Second,
			rps:      10,
			expected: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.delay, tt.rps)
			assert.Equal(t, tt.expected, rl.Interval())
		})
	}
}
