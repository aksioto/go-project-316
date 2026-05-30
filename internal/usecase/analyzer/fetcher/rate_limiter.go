package fetcher

import (
	"time"

	"golang.org/x/time/rate"
)

func NewRateLimiter(delay time.Duration, rps float64) *rate.Limiter {
	if rps > 0 {
		return rate.NewLimiter(rate.Limit(rps), 1)
	}

	if delay > 0 {
		return rate.NewLimiter(rate.Every(delay), 1)
	}

	return rate.NewLimiter(rate.Inf, 0)
}
