package analyzer

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"code/internal/domain"

	"go.uber.org/zap"
)

const (
	defaultRetryDelay = 100 * time.Millisecond
)

type RetryFetcher struct {
	logger     *zap.Logger
	fetcher    Fetcher
	maxRetries int
	retryDelay time.Duration
}

func NewRetryFetcher(logger *zap.Logger, fetcher Fetcher, maxRetries int) *RetryFetcher {
	return &RetryFetcher{
		logger:     logger,
		fetcher:    fetcher,
		maxRetries: maxRetries,
		retryDelay: defaultRetryDelay,
	}
}

func (r *RetryFetcher) Fetch(ctx context.Context, url string) (domain.FetchResult, error) {
	var lastResult domain.FetchResult
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if ctx.Err() != nil {
			return domain.FetchResult{}, ctx.Err()
		}

		if attempt > 0 {
			r.logger.Debug("retrying request",
				zap.String("url", url),
				zap.Int("attempt", attempt),
				zap.Int("max_retries", r.maxRetries),
			)

			select {
			case <-ctx.Done():
				return domain.FetchResult{}, ctx.Err()
			case <-time.After(r.retryDelay):
			}
		}

		result, err := r.fetcher.Fetch(ctx, url)
		lastResult = result
		lastErr = err

		if err != nil {
			if r.isRetryableError(err) {
				r.logger.Debug("retryable error",
					zap.String("url", url),
					zap.Error(err),
				)
				continue
			}
			return result, err
		}

		if r.isRetryableStatusCode(result.StatusCode) {
			r.logger.Debug("retryable status code",
				zap.String("url", url),
				zap.Int("status", result.StatusCode),
			)
			continue
		}

		return result, nil
	}

	return lastResult, lastErr
}

func (r *RetryFetcher) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	var opErr *net.OpError
	return errors.As(err, &opErr)
}

func (r *RetryFetcher) isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}
