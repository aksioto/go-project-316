package analyzer

import (
	"context"
	"time"
)

type RateLimiter struct {
	interval time.Duration
	lastTime time.Time
}

func NewRateLimiter(delay time.Duration, rps float64) *RateLimiter {
	interval := delay

	if rps > 0 {
		interval = time.Duration(float64(time.Second) / rps)
	}

	return &RateLimiter{
		interval: interval,
	}
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	if r.interval <= 0 {
		return nil
	}

	if r.lastTime.IsZero() {
		r.lastTime = time.Now()
		return nil
	}

	elapsed := time.Since(r.lastTime)
	if elapsed >= r.interval {
		r.lastTime = time.Now()
		return nil
	}

	waitTime := r.interval - elapsed

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		r.lastTime = time.Now()
		return nil
	}
}

func (r *RateLimiter) Interval() time.Duration {
	return r.interval
}
