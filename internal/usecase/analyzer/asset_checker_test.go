package analyzer

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"code/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAssetChecker_ExtractsAssets(t *testing.T) {
	html := []byte(`
		<html>
		<head>
			<link rel="stylesheet" href="/styles.css">
			<script src="/app.js"></script>
		</head>
		<body>
			<img src="/logo.png">
		</body>
		</html>
	`)

	fetcher := &mockAssetFetcher{
		responses: map[string]domain.FetchResult{
			"http://example.com/styles.css": {StatusCode: 200, ContentLength: 1000},
			"http://example.com/app.js":     {StatusCode: 200, ContentLength: 2000},
			"http://example.com/logo.png":   {StatusCode: 200, ContentLength: 5000},
		},
	}

	checker := NewAssetChecker(
		zap.NewNop(),
		fetcher,
		NewRateLimiter(0, 0),
		NewAssetExtractor(),
	)

	assets := checker.Check(context.Background(), "http://example.com/page", html)

	require.Len(t, assets, 3)

	assetMap := make(map[string]domain.Asset)
	for _, a := range assets {
		assetMap[a.URL] = a
	}

	css := assetMap["http://example.com/styles.css"]
	assert.Equal(t, domain.AssetStyle, css.Type)
	assert.Equal(t, int64(1000), css.SizeBytes)
	assert.Equal(t, 200, css.StatusCode)

	js := assetMap["http://example.com/app.js"]
	assert.Equal(t, domain.AssetScript, js.Type)
	assert.Equal(t, int64(2000), js.SizeBytes)

	img := assetMap["http://example.com/logo.png"]
	assert.Equal(t, domain.AssetImage, img.Type)
	assert.Equal(t, int64(5000), img.SizeBytes)
}

func TestAssetChecker_CachesAssets(t *testing.T) {
	var callCount int32

	fetcher := &mockAssetFetcher{
		responses: map[string]domain.FetchResult{
			"http://example.com/shared.js": {StatusCode: 200, ContentLength: 100},
		},
		onFetch: func(url string) {
			atomic.AddInt32(&callCount, 1)
		},
	}

	checker := NewAssetChecker(
		zap.NewNop(),
		fetcher,
		NewRateLimiter(0, 0),
		NewAssetExtractor(),
	)

	html1 := []byte(`<script src="/shared.js"></script>`)
	html2 := []byte(`<script src="/shared.js"></script>`)

	assets1 := checker.Check(context.Background(), "http://example.com/page1", html1)
	assets2 := checker.Check(context.Background(), "http://example.com/page2", html2)

	require.Len(t, assets1, 1)
	require.Len(t, assets2, 1)

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "asset should be fetched only once (cached)")
	assert.Equal(t, assets1[0].SizeBytes, assets2[0].SizeBytes)
}

func TestAssetChecker_NoContentLength(t *testing.T) {
	bodyContent := []byte("body content without Content-Length header")

	fetcher := &mockAssetFetcher{
		responses: map[string]domain.FetchResult{
			"http://example.com/style.css": {
				StatusCode:    200,
				ContentLength: -1,
				Body:          bodyContent,
			},
		},
	}

	checker := NewAssetChecker(
		zap.NewNop(),
		fetcher,
		NewRateLimiter(0, 0),
		NewAssetExtractor(),
	)

	html := []byte(`<link rel="stylesheet" href="/style.css">`)
	assets := checker.Check(context.Background(), "http://example.com/page", html)

	require.Len(t, assets, 1)
	assert.Equal(t, int64(len(bodyContent)), assets[0].SizeBytes, "size should be calculated from body")
	assert.Empty(t, assets[0].Error)
}

func TestAssetChecker_ErrorStatus(t *testing.T) {
	fetcher := &mockAssetFetcher{
		responses: map[string]domain.FetchResult{
			"http://example.com/missing.png": {StatusCode: 404},
		},
	}

	checker := NewAssetChecker(
		zap.NewNop(),
		fetcher,
		NewRateLimiter(0, 0),
		NewAssetExtractor(),
	)

	html := []byte(`<img src="/missing.png">`)
	assets := checker.Check(context.Background(), "http://example.com/page", html)

	require.Len(t, assets, 1)
	assert.Equal(t, 404, assets[0].StatusCode)
	assert.Equal(t, "HTTP 404", assets[0].Error)
	assert.Equal(t, int64(0), assets[0].SizeBytes)
}

func TestAssetChecker_ServerError(t *testing.T) {
	fetcher := &mockAssetFetcher{
		responses: map[string]domain.FetchResult{
			"http://example.com/broken.js": {StatusCode: 500},
		},
	}

	checker := NewAssetChecker(
		zap.NewNop(),
		fetcher,
		NewRateLimiter(0, 0),
		NewAssetExtractor(),
	)

	html := []byte(`<script src="/broken.js"></script>`)
	assets := checker.Check(context.Background(), "http://example.com/page", html)

	require.Len(t, assets, 1)
	assert.Equal(t, 500, assets[0].StatusCode)
	assert.Equal(t, "HTTP 500", assets[0].Error)
}

func TestAssetChecker_ContextCancellation(t *testing.T) {
	fetcher := &mockAssetFetcher{
		responses: map[string]domain.FetchResult{
			"http://example.com/slow.js": {StatusCode: 200, ContentLength: 100},
		},
		delay: 100 * time.Millisecond,
	}

	checker := NewAssetChecker(
		zap.NewNop(),
		fetcher,
		NewRateLimiter(0, 0),
		NewAssetExtractor(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	html := []byte(`<script src="/slow.js"></script>`)
	assets := checker.Check(ctx, "http://example.com/page", html)

	assert.Empty(t, assets, "should stop on context cancellation")
}

type mockAssetFetcher struct {
	responses map[string]domain.FetchResult
	onFetch   func(url string)
	delay     time.Duration
}

func (m *mockAssetFetcher) Fetch(ctx context.Context, url string) (domain.FetchResult, error) {
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return domain.FetchResult{}, ctx.Err()
		case <-time.After(m.delay):
		}
	}

	if m.onFetch != nil {
		m.onFetch(url)
	}

	if result, ok := m.responses[url]; ok {
		return result, nil
	}
	return domain.FetchResult{StatusCode: http.StatusNotFound}, nil
}
