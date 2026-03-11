package analyzer

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"code/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockFetcher struct {
	responses []mockResponse
	callCount int32
}

type mockResponse struct {
	result domain.FetchResult
	err    error
}

func (m *mockFetcher) Fetch(ctx context.Context, url string) (domain.FetchResult, error) {
	idx := int(atomic.AddInt32(&m.callCount, 1)) - 1
	if idx >= len(m.responses) {
		return domain.FetchResult{}, errors.New("no more responses")
	}
	return m.responses[idx].result, m.responses[idx].err
}

func (m *mockFetcher) CallCount() int {
	return int(atomic.LoadInt32(&m.callCount))
}

func TestRetryFetcher_NoRetryOnSuccess(t *testing.T) {
	mock := &mockFetcher{
		responses: []mockResponse{
			{result: domain.FetchResult{StatusCode: http.StatusOK, Body: []byte("ok")}},
		},
	}

	rf := NewRetryFetcher(zap.NewNop(), mock, 3)
	result, err := rf.Fetch(context.Background(), "http://example.com")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Equal(t, 1, mock.CallCount(), "should not retry on success")
}

func TestRetryFetcher_RetryOnServerError(t *testing.T) {
	mock := &mockFetcher{
		responses: []mockResponse{
			{result: domain.FetchResult{StatusCode: http.StatusInternalServerError}},
			{result: domain.FetchResult{StatusCode: http.StatusOK, Body: []byte("ok")}},
		},
	}

	rf := NewRetryFetcher(zap.NewNop(), mock, 3)
	rf.retryDelay = 1 * time.Millisecond

	result, err := rf.Fetch(context.Background(), "http://example.com")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Equal(t, 2, mock.CallCount(), "should retry once after 500")
}

func TestRetryFetcher_RetryOn429(t *testing.T) {
	mock := &mockFetcher{
		responses: []mockResponse{
			{result: domain.FetchResult{StatusCode: http.StatusTooManyRequests}},
			{result: domain.FetchResult{StatusCode: http.StatusTooManyRequests}},
			{result: domain.FetchResult{StatusCode: http.StatusOK, Body: []byte("ok")}},
		},
	}

	rf := NewRetryFetcher(zap.NewNop(), mock, 3)
	rf.retryDelay = 1 * time.Millisecond

	result, err := rf.Fetch(context.Background(), "http://example.com")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Equal(t, 3, mock.CallCount(), "should retry twice after 429")
}

func TestRetryFetcher_ExhaustedRetries(t *testing.T) {
	mock := &mockFetcher{
		responses: []mockResponse{
			{result: domain.FetchResult{StatusCode: http.StatusServiceUnavailable}},
			{result: domain.FetchResult{StatusCode: http.StatusServiceUnavailable}},
			{result: domain.FetchResult{StatusCode: http.StatusServiceUnavailable}},
		},
	}

	rf := NewRetryFetcher(zap.NewNop(), mock, 2)
	rf.retryDelay = 1 * time.Millisecond

	result, err := rf.Fetch(context.Background(), "http://example.com")

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, result.StatusCode)
	assert.Equal(t, 3, mock.CallCount(), "should make retries+1 attempts")
}

func TestRetryFetcher_NoRetryOn404(t *testing.T) {
	mock := &mockFetcher{
		responses: []mockResponse{
			{result: domain.FetchResult{StatusCode: http.StatusNotFound}},
		},
	}

	rf := NewRetryFetcher(zap.NewNop(), mock, 3)
	result, err := rf.Fetch(context.Background(), "http://example.com")

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, result.StatusCode)
	assert.Equal(t, 1, mock.CallCount(), "should not retry on 404")
}

func TestRetryFetcher_RetryOnNetworkError(t *testing.T) {
	networkErr := &mockNetError{temporary: true}
	mock := &mockFetcher{
		responses: []mockResponse{
			{err: networkErr},
			{result: domain.FetchResult{StatusCode: http.StatusOK}},
		},
	}

	rf := NewRetryFetcher(zap.NewNop(), mock, 3)
	rf.retryDelay = 1 * time.Millisecond

	result, err := rf.Fetch(context.Background(), "http://example.com")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Equal(t, 2, mock.CallCount(), "should retry on network error")
}

func TestRetryFetcher_ContextCancellation(t *testing.T) {
	mock := &mockFetcher{
		responses: []mockResponse{
			{result: domain.FetchResult{StatusCode: http.StatusServiceUnavailable}},
			{result: domain.FetchResult{StatusCode: http.StatusServiceUnavailable}},
			{result: domain.FetchResult{StatusCode: http.StatusServiceUnavailable}},
		},
	}

	rf := NewRetryFetcher(zap.NewNop(), mock, 5)
	rf.retryDelay = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := rf.Fetch(ctx, "http://example.com")

	assert.ErrorIs(t, err, context.Canceled)
	assert.LessOrEqual(t, mock.CallCount(), 2, "should stop retrying on context cancel")
}

func TestRetryFetcher_MaxAttemptsIsRetriesPlusOne(t *testing.T) {
	testCases := []struct {
		name     string
		retries  int
		expected int
	}{
		{"retries=0", 0, 1},
		{"retries=1", 1, 2},
		{"retries=2", 2, 3},
		{"retries=5", 5, 6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			responses := make([]mockResponse, tc.expected)
			for i := range responses {
				responses[i] = mockResponse{result: domain.FetchResult{StatusCode: http.StatusServiceUnavailable}}
			}

			mock := &mockFetcher{responses: responses}
			rf := NewRetryFetcher(zap.NewNop(), mock, tc.retries)
			rf.retryDelay = 1 * time.Millisecond

			_, _ = rf.Fetch(context.Background(), "http://example.com")

			assert.Equal(t, tc.expected, mock.CallCount())
		})
	}
}

type mockNetError struct {
	temporary bool
	timeout   bool
}

func (e *mockNetError) Error() string   { return "mock network error" }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return e.temporary }
