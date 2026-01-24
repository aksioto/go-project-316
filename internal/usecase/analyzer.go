package usecase

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

type Analyzer struct {
	fetcher Fetcher
	opts    Options
}

func NewAnalyzer(fetcher Fetcher, opts Options) *Analyzer {
	return &Analyzer{
		fetcher: fetcher,
		opts:    opts,
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
	// TODO: SEO, Assets, BrokenLinks

	return page
}
