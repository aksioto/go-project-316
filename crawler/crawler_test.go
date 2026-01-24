package crawler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code/crawler"
	"code/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><head><title>Test</title></head><body><h1>Hello</h1></body></html>"))
	}))
	defer server.Close()

	tests := []struct {
		name           string
		httpClient     *http.Client
		url            string
		depth          int
		wantStatus     string
		wantHTTPStatus int
		wantError      bool
		ctxFunc        func() (context.Context, context.CancelFunc)
	}{
		{
			name:           "successful fetch",
			httpClient:     server.Client(),
			url:            server.URL,
			depth:          1,
			wantStatus:     "ok",
			wantHTTPStatus: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "network error",
			httpClient:     testutils.NewErrorClient(errors.New("connection refused")),
			url:            "http://invalid.localhost.test:99999",
			depth:          1,
			wantStatus:     "error",
			wantHTTPStatus: 0,
			wantError:      true,
		},
		{
			name:           "timeout",
			httpClient:     testutils.NewTimeoutClient(),
			url:            "http://example.com",
			depth:          1,
			wantStatus:     "error",
			wantHTTPStatus: 0,
			wantError:      true,
			ctxFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 10*time.Millisecond)
			},
		},
		{
			name:           "server error 500",
			httpClient:     testutils.NewResponseClient(http.StatusInternalServerError, "Internal Server Error"),
			url:            "http://example.com",
			depth:          2,
			wantStatus:     "ok",
			wantHTTPStatus: http.StatusInternalServerError,
			wantError:      false,
		},
		{
			name:           "not found 404",
			httpClient:     testutils.NewResponseClient(http.StatusNotFound, "Not Found"),
			url:            "http://example.com/missing",
			depth:          1,
			wantStatus:     "ok",
			wantHTTPStatus: http.StatusNotFound,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.ctxFunc != nil {
				var cancel context.CancelFunc
				ctx, cancel = tt.ctxFunc()
				t.Cleanup(cancel)
			}

			opts := testutils.NewCrawlerOptions(tt.url, tt.depth, tt.httpClient)
			result, err := crawler.Analyze(ctx, opts)
			require.NoError(t, err)
			require.NotEmpty(t, result)

			report, err := testutils.ParseReport(result)
			require.NoError(t, err)

			assert.Equal(t, tt.url, report.RootURL)
			assert.Equal(t, tt.depth, report.Depth)
			require.Len(t, report.Pages, 1)

			page := report.Pages[0]
			assert.Equal(t, tt.url, page.URL)
			assert.Equal(t, tt.wantStatus, page.Status)
			assert.Equal(t, tt.wantHTTPStatus, page.HTTPStatus)

			if tt.wantError {
				assert.NotEmpty(t, page.Error)
			} else {
				assert.Empty(t, page.Error)
			}
		})
	}
}

func TestAnalyzeRequiresHTTPClient(t *testing.T) {
	opts := testutils.NewCrawlerOptions("https://example.com", 1, nil)

	_, err := crawler.Analyze(context.Background(), opts)
	require.Error(t, err)
}
