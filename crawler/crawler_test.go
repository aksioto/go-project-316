package crawler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"code/crawler"
	"code/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	const successHTML = "<html><head><title>Test</title></head><body><h1>Hello</h1></body></html>"

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
			httpClient:     testutils.NewResponseClient(http.StatusOK, successHTML),
			url:            "http://example.com/",
			depth:          1,
			wantStatus:     "ok",
			wantHTTPStatus: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "network error",
			httpClient:     testutils.NewErrorClient(errors.New("connection refused")),
			url:            "http://invalid.localhost.test:99999/",
			depth:          1,
			wantStatus:     "error",
			wantHTTPStatus: 0,
			wantError:      true,
		},
		{
			name:           "timeout",
			httpClient:     testutils.NewTimeoutClient(),
			url:            "http://example.com/",
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
			url:            "http://example.com/",
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

func TestAnalyzeBrokenLinks(t *testing.T) {
	const pageURL = "http://example.com/page"
	const htmlBody = `<!doctype html>
<html>
  <body>
    <a href="http://example.com/ok">OK</a>
    <a href="/broken">Broken</a>
    <a href="">Empty</a>
    <a href="mailto:test@example.com">Mail</a>
  </body>
</html>`

	client := testutils.NewStubClient(map[string]testutils.StubResponse{
		pageURL: {
			StatusCode: http.StatusOK,
			Body:       htmlBody,
		},
		"http://example.com/ok": {
			StatusCode: http.StatusOK,
			Body:       "ok",
		},
		"http://example.com/broken": {
			StatusCode: http.StatusNotFound,
			Body:       "not found",
		},
	})

	opts := testutils.NewCrawlerOptions(pageURL, 1, client)
	result, err := crawler.Analyze(context.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	report, err := testutils.ParseReport(result)
	require.NoError(t, err)
	require.Len(t, report.Pages, 1)

	page := report.Pages[0]
	require.Len(t, page.BrokenLinks, 1)

	broken := page.BrokenLinks[0]
	assert.Equal(t, "http://example.com/broken", broken.URL)
	assert.Equal(t, http.StatusNotFound, broken.StatusCode)
	assert.Empty(t, broken.Error)
}

func TestAnalyzeSEOWithTags(t *testing.T) {
	const pageURL = "http://example.com/seo"
	const htmlBody = `<!doctype html>
<html>
  <head>
    <title>Hello &amp; World</title>
    <meta name="description" content="Best &amp; reliable" />
  </head>
  <body>
    <h1>Landing</h1>
  </body>
</html>`

	client := testutils.NewStubClient(map[string]testutils.StubResponse{
		pageURL: {
			StatusCode: http.StatusOK,
			Body:       htmlBody,
		},
	})

	opts := testutils.NewCrawlerOptions(pageURL, 1, client)
	result, err := crawler.Analyze(context.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	report, err := testutils.ParseReport(result)
	require.NoError(t, err)
	require.Len(t, report.Pages, 1)

	page := report.Pages[0]
	require.NotNil(t, page.SEO)
	assert.True(t, page.SEO.HasTitle)
	assert.Equal(t, "Hello & World", page.SEO.Title)
	assert.True(t, page.SEO.HasDescription)
	assert.Equal(t, "Best & reliable", page.SEO.Description)
	assert.True(t, page.SEO.HasH1)
}

func TestAnalyzeSEOMissingTags(t *testing.T) {
	const pageURL = "http://example.com/seo-empty"
	const htmlBody = `<!doctype html>
<html>
  <head></head>
  <body>
    <p>Content</p>
  </body>
</html>`

	client := testutils.NewStubClient(map[string]testutils.StubResponse{
		pageURL: {
			StatusCode: http.StatusOK,
			Body:       htmlBody,
		},
	})

	opts := testutils.NewCrawlerOptions(pageURL, 1, client)
	result, err := crawler.Analyze(context.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	report, err := testutils.ParseReport(result)
	require.NoError(t, err)
	require.Len(t, report.Pages, 1)

	page := report.Pages[0]
	require.NotNil(t, page.SEO)
	assert.False(t, page.SEO.HasTitle)
	assert.Empty(t, page.SEO.Title)
	assert.False(t, page.SEO.HasDescription)
	assert.Empty(t, page.SEO.Description)
	assert.False(t, page.SEO.HasH1)
}

func TestAnalyzeDepthLimit(t *testing.T) {
	const rootURL = "http://example.com/"
	const childURL = "http://example.com/child"
	const htmlBody = `<!doctype html>
<html>
  <head><title>Root</title></head>
  <body>
    <a href="/child">Child</a>
  </body>
</html>`

	client := testutils.NewStubClient(map[string]testutils.StubResponse{
		rootURL: {
			StatusCode: http.StatusOK,
			Body:       htmlBody,
		},
		childURL: {
			StatusCode: http.StatusOK,
			Body:       "<html><head><title>Child</title></head><body><h1>Child</h1></body></html>",
		},
	})

	opts := testutils.NewCrawlerOptions(rootURL, 1, client)
	result, err := crawler.Analyze(context.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	report, err := testutils.ParseReport(result)
	require.NoError(t, err)
	require.Len(t, report.Pages, 1)
	assert.Equal(t, rootURL, report.Pages[0].URL)
}

func TestAnalyzeDepthInternalLinks(t *testing.T) {
	const rootURL = "http://example.com/"
	const internalOne = "http://example.com/internal-1"
	const internalTwo = "http://example.com/internal-2"
	const external = "http://external.com/page"
	const htmlBody = `<!doctype html>
<html>
  <head><title>Root</title></head>
  <body>
    <a href="/internal-1">Internal 1</a>
    <a href="/internal-2">Internal 2</a>
    <a href="/internal-1">Duplicate</a>
    <a href="http://external.com/page">External</a>
  </body>
</html>`

	client := testutils.NewStubClient(map[string]testutils.StubResponse{
		rootURL: {
			StatusCode: http.StatusOK,
			Body:       htmlBody,
		},
		internalOne: {
			StatusCode: http.StatusOK,
			Body:       "<html><head><title>Internal One</title><meta name=\"description\" content=\"First\"/></head><body><h1>One</h1></body></html>",
		},
		internalTwo: {
			StatusCode: http.StatusOK,
			Body:       "<html><head><title>Internal Two</title></head><body><h1>Two</h1></body></html>",
		},
		external: {
			StatusCode: http.StatusOK,
			Body:       "<html><head><title>External</title></head><body><h1>External</h1></body></html>",
		},
	})

	opts := testutils.NewCrawlerOptions(rootURL, 2, client)
	result, err := crawler.Analyze(context.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	report, err := testutils.ParseReport(result)
	require.NoError(t, err)
	require.Len(t, report.Pages, 3)

	rootPage := findPage(report.Pages, rootURL)
	internalPageOne := findPage(report.Pages, internalOne)
	internalPageTwo := findPage(report.Pages, internalTwo)

	require.NotNil(t, rootPage)
	require.NotNil(t, internalPageOne)
	require.NotNil(t, internalPageTwo)
	assert.Nil(t, findPage(report.Pages, external))
	assert.Equal(t, 1, countPages(report.Pages, internalOne))

	require.NotNil(t, internalPageOne.SEO)
	assert.True(t, internalPageOne.SEO.HasTitle)
	assert.Equal(t, "Internal One", internalPageOne.SEO.Title)
	assert.True(t, internalPageOne.SEO.HasDescription)
	assert.Equal(t, "First", internalPageOne.SEO.Description)
	assert.True(t, internalPageOne.SEO.HasH1)

	require.NotNil(t, internalPageTwo.SEO)
	assert.True(t, internalPageTwo.SEO.HasTitle)
	assert.Equal(t, "Internal Two", internalPageTwo.SEO.Title)
	assert.False(t, internalPageTwo.SEO.HasDescription)
	assert.Empty(t, internalPageTwo.SEO.Description)
	assert.True(t, internalPageTwo.SEO.HasH1)
}

func TestAnalyzeDeduplicatesIndexAliases(t *testing.T) {
	const rootURL = "http://example.com/"
	const htmlBody = `<!doctype html>
<html>
  <head><title>Root</title></head>
  <body>
    <a href="/index.html">Home via index.html</a>
    <a href="/about/">About</a>
    <a href="/about/index.html">About via index.html</a>
  </body>
</html>`

	client := testutils.NewStubClient(map[string]testutils.StubResponse{
		rootURL: {
			StatusCode: http.StatusOK,
			Body:       htmlBody,
		},
		"http://example.com/about/": {
			StatusCode: http.StatusOK,
			Body:       "<html><head><title>About</title></head><body><h1>About</h1></body></html>",
		},
	})

	opts := testutils.NewCrawlerOptions(rootURL, 3, client)
	result, err := crawler.Analyze(context.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	report, err := testutils.ParseReport(result)
	require.NoError(t, err)

	require.Len(t, report.Pages, 2, "should have only root and about, no duplicates")
	assert.Equal(t, 1, countPages(report.Pages, rootURL), "root should appear once")
	assert.Equal(t, 1, countPages(report.Pages, "http://example.com/about/"), "about should appear once")
}

func TestAnalyzeSkipsNonHTMLPages(t *testing.T) {
	const rootURL = "http://example.com/"
	const childURL = "http://example.com/child"
	const cssURL = "http://example.com/style.css"
	const htmlBody = `<!doctype html>
<html>
  <head>
    <title>Root</title>
    <link rel="stylesheet" href="/style.css" />
  </head>
  <body>
    <a href="/child">Child</a>
  </body>
</html>`

	client := testutils.NewStubClient(map[string]testutils.StubResponse{
		rootURL: {
			StatusCode: http.StatusOK,
			Body:       htmlBody,
		},
		childURL: {
			StatusCode: http.StatusOK,
			Body:       "<html><head><title>Child</title></head><body><h1>Child</h1></body></html>",
		},
		cssURL: {
			StatusCode:  http.StatusOK,
			Body:        "body { color: black; }",
			ContentType: "text/css; charset=utf-8",
		},
	})

	opts := testutils.NewCrawlerOptions(rootURL, 2, client)
	result, err := crawler.Analyze(context.Background(), opts)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	report, err := testutils.ParseReport(result)
	require.NoError(t, err)
	require.Len(t, report.Pages, 2)
	assert.Nil(t, findPage(report.Pages, cssURL))
	require.NotNil(t, findPage(report.Pages, childURL))
}

func TestAnalyzeRequiresHTTPClient(t *testing.T) {
	opts := testutils.NewCrawlerOptions("https://example.com", 1, nil)

	_, err := crawler.Analyze(context.Background(), opts)
	require.Error(t, err)
}

func findPage(pages []testutils.Page, url string) *testutils.Page {
	for i := range pages {
		if pages[i].URL == url {
			return &pages[i]
		}
	}
	return nil
}

func countPages(pages []testutils.Page, url string) int {
	count := 0
	for _, page := range pages {
		if page.URL == url {
			count++
		}
	}
	return count
}
