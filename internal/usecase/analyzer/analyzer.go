package analyzer

import (
	"context"
	"net/url"
	"time"

	"code/internal/domain"

	"go.uber.org/zap"
)

type Options struct {
	URL         string
	Depth       int
	Retries     int
	Delay       time.Duration
	Timeout     time.Duration
	RPS         float64
	UserAgent   string
	Concurrency int
}

type Fetcher interface {
	Fetch(ctx context.Context, url string) (domain.FetchResult, error)
}

type Analyzer struct {
	logger            *zap.Logger
	pageFetcher       PageFetcher
	domainFilter      DomainFilter
	linkExtractor     *LinkExtractor
	brokenLinkChecker *BrokenLinkChecker
	seoAnalyzer       *SEOAnalyzer
	assetChecker      *AssetChecker
	rateLimiter       *RateLimiter
	opts              Options
}

type queueItem struct {
	url   string
	depth int
}

func NewDefaultAnalyzer(logger *zap.Logger, fetcher Fetcher, opts Options) *Analyzer {
	if logger == nil {
		logger = zap.NewNop()
	}

	retryFetcher := NewRetryFetcher(logger, fetcher, opts.Retries)
	contentTypeFilter := NewContentTypeFilter()
	pageFetcher := NewPageFetcher(logger, retryFetcher, contentTypeFilter)
	rateLimiter := NewRateLimiter(opts.Delay, opts.RPS)
	assetExtractor := NewAssetExtractor()

	return &Analyzer{
		logger:            logger,
		pageFetcher:       pageFetcher,
		domainFilter:      NewDomainFilter(),
		linkExtractor:     NewLinkExtractor(),
		brokenLinkChecker: NewBrokenLinkChecker(logger, retryFetcher, rateLimiter),
		seoAnalyzer:       NewSEOAnalyzer(logger),
		assetChecker:      NewAssetChecker(logger, retryFetcher, rateLimiter, assetExtractor),
		rateLimiter:       rateLimiter,
		opts:              opts,
	}
}

func (a *Analyzer) Analyze(ctx context.Context) domain.Report {
	a.logger.Debug("starting analysis",
		zap.String("url", a.opts.URL),
		zap.Int("max_depth", a.opts.Depth),
	)

	report := domain.Report{
		RootURL:     a.opts.URL,
		MaxDepth:    a.opts.Depth,
		GeneratedAt: time.Now(),
		Pages:       []domain.Page{},
	}

	root, err := url.Parse(a.opts.URL)
	if err != nil {
		a.logger.Debug("failed to parse root URL", zap.Error(err))
		result := a.pageFetcher.Fetch(ctx, a.opts.URL, 0)
		report.Pages = append(report.Pages, result.Page)
		return report
	}

	startURL := NormalizeURL(a.opts.URL)
	queue := []queueItem{{url: startURL, depth: 0}}
	visited := map[string]struct{}{startURL: {}}

	for len(queue) > 0 {
		if ctx.Err() != nil {
			break
		}

		item := queue[0]
		queue = queue[1:]

		if err := a.rateLimiter.Wait(ctx); err != nil {
			break
		}

		page, links, shouldAdd := a.processPage(ctx, item)
		if shouldAdd {
			report.Pages = append(report.Pages, page)
		}

		if item.depth+1 < a.opts.Depth {
			queue = a.enqueueLinks(root, links, item.depth+1, visited, queue)
		}
	}

	a.logger.Debug("analysis completed",
		zap.Int("pages_found", len(report.Pages)),
	)

	return report
}

func (a *Analyzer) processPage(ctx context.Context, item queueItem) (domain.Page, []domain.Link, bool) {
	result := a.pageFetcher.Fetch(ctx, item.url, item.depth)
	page := result.Page

	isStartPage := item.depth == 0
	hasError := page.Err != nil || (page.StatusCode >= 400 && page.StatusCode < 600)

	if !result.IsHTML {
		return page, nil, isStartPage && hasError
	}

	links := a.linkExtractor.Extract(item.url, result.Body)
	page.BrokenLinks = a.brokenLinkChecker.Check(ctx, links)
	page.Assets = a.assetChecker.Check(ctx, item.url, result.Body)
	seo := a.seoAnalyzer.Analyze(result.Body)
	page.SEO = &seo

	return page, links, true
}

func (a *Analyzer) enqueueLinks(
	root *url.URL,
	links []domain.Link,
	depth int,
	visited map[string]struct{},
	queue []queueItem,
) []queueItem {
	for _, link := range links {
		normalizedURL := NormalizeURL(link.URL)

		if !a.domainFilter.IsSameDomain(root, normalizedURL) {
			continue
		}
		if _, seen := visited[normalizedURL]; seen {
			continue
		}

		visited[normalizedURL] = struct{}{}
		queue = append(queue, queueItem{url: normalizedURL, depth: depth})
	}
	return queue
}
