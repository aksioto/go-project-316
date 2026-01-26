package analyzer

import (
	"context"
	"time"

	"code/internal/domain"
)

type Options struct {
	URL         string
	Depth       int
	Retries     int
	Delay       time.Duration
	Timeout     time.Duration
	UserAgent   string
	Concurrency int
}

type Fetcher interface {
	Fetch(ctx context.Context, url string) (domain.FetchResult, error)
}

type LinkExtractor interface {
	Extract(baseURL string, body []byte) []domain.Link
}

type BrokenLinkChecker interface {
	Check(ctx context.Context, links []domain.Link) []domain.BrokenLink
}

type SEOAnalyzer interface {
	Analyze(body []byte) domain.SEOResult
}

type Analyzer struct {
	fetcher           Fetcher
	opts              Options
	linkExtractor     LinkExtractor
	brokenLinkChecker BrokenLinkChecker
	seoAnalyzer       SEOAnalyzer
}

func NewAnalyzer(fetcher Fetcher, extractor LinkExtractor, checker BrokenLinkChecker, seoAnalyzer SEOAnalyzer, opts Options) *Analyzer {
	return &Analyzer{
		fetcher:           fetcher,
		linkExtractor:     extractor,
		brokenLinkChecker: checker,
		seoAnalyzer:       seoAnalyzer,
		opts:              opts,
	}
}

func (a *Analyzer) Analyze(ctx context.Context) domain.Report {
	report := domain.Report{
		RootURL:     a.opts.URL,
		MaxDepth:    a.opts.Depth,
		GeneratedAt: time.Now(),
		Pages:       []domain.Page{},
	}

	page := a.fetchPage(ctx, a.opts.URL, 0)
	report.Pages = append(report.Pages, page)

	return report
}

func (a *Analyzer) fetchPage(ctx context.Context, url string, depth int) domain.Page {
	page := domain.Page{
		URL:   url,
		Depth: depth,
	}

	result, err := a.fetcher.Fetch(ctx, url)
	if err != nil {
		page.Err = err
		return page
	}

	page.StatusCode = result.StatusCode
	links := a.linkExtractor.Extract(url, result.Body)
	page.BrokenLinks = a.brokenLinkChecker.Check(ctx, links)
	seo := a.seoAnalyzer.Analyze(result.Body)
	page.SEO = &seo
	// TODO: SEO, Assets

	return page
}
