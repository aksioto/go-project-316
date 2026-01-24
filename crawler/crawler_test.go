package crawler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"code/crawler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testReport struct {
	RootURL string     `json:"root_url"`
	Depth   int        `json:"depth"`
	Pages   []testPage `json:"pages"`
}

type testPage struct {
	URL        string `json:"url"`
	HTTPStatus int    `json:"http_status"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

func parseReport(data []byte) (*testReport, error) {
	var report testReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newErrorClient(err error) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, err
		}),
	}
}

func newResponseClient(status int, body string) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}
}

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
			httpClient:     newErrorClient(errors.New("connection refused")),
			url:            "http://invalid.localhost.test:99999",
			depth:          1,
			wantStatus:     "error",
			wantHTTPStatus: 0,
			wantError:      true,
		},
		{
			name:           "server error 500",
			httpClient:     newResponseClient(http.StatusInternalServerError, "Internal Server Error"),
			url:            "http://example.com",
			depth:          2,
			wantStatus:     "ok",
			wantHTTPStatus: http.StatusInternalServerError,
			wantError:      false,
		},
		{
			name:           "not found 404",
			httpClient:     newResponseClient(http.StatusNotFound, "Not Found"),
			url:            "http://example.com/missing",
			depth:          1,
			wantStatus:     "ok",
			wantHTTPStatus: http.StatusNotFound,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := crawler.Options{
				URL:        tt.url,
				Depth:      tt.depth,
				Timeout:    5 * time.Second,
				HTTPClient: tt.httpClient,
				IndentJSON: false,
			}

			result, err := crawler.Analyze(context.Background(), opts)
			require.NoError(t, err)
			require.NotEmpty(t, result)

			report, err := parseReport(result)
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
	opts := crawler.Options{
		URL:        "https://example.com",
		Depth:      1,
		IndentJSON: false,
	}

	_, err := crawler.Analyze(context.Background(), opts)
	require.Error(t, err)
}
