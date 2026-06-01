package analyzer

import (
	"context"
	"net/url"
	"time"

	"code/internal/domain"
	"code/internal/usecase/analyzer/extractor"
	"code/internal/usecase/analyzer/fetcher"
	"code/internal/usecase/analyzer/filter"

	"go.uber.org/zap"
)

type Analyzer struct {
	logger       *zap.Logger
	processor    *PageProcessor
	pageFetcher  PageFetcher
	domainFilter filter.DomainFilter
	rateLimiter  Waiter
	opts         Options
}

func NewDefaultAnalyzer(logger *zap.Logger, f fetcher.Fetcher, opts Options) *Analyzer {
	if logger == nil {
		logger = zap.NewNop()
	}

	retryFetcher := fetcher.NewRetryFetcher(logger, f, opts.Retries)
	contentTypeFilter := filter.NewContentTypeFilter()
	pf := NewPageFetcher(logger, retryFetcher, contentTypeFilter)
	rateLimiter := fetcher.NewRateLimiter(opts.Delay, opts.RPS)

	linkExtractor := extractor.NewLinkExtractor()
	assetExtractor := extractor.NewAssetExtractor()

	domainFilter := filter.NewDomainFilter()

	enricher := NewCompositeEnricher(
		NewSEOEnricher(logger),
		NewLinksEnricher(logger, retryFetcher, rateLimiter, linkExtractor),
		NewAssetsEnricher(logger, retryFetcher, rateLimiter, assetExtractor),
	)

	processor := NewPageProcessor(logger, pf, linkExtractor, enricher)

	return &Analyzer{
		logger:       logger,
		processor:    processor,
		pageFetcher:  pf,
		domainFilter: domainFilter,
		rateLimiter:  rateLimiter,
		opts:         opts,
	}
}

func (a *Analyzer) Analyze(ctx context.Context) domain.Report {
	a.logger.Debug("starting analysis",
		zap.String("url", a.opts.URL),
		zap.Int("max_depth", a.opts.Depth),
		zap.Int("workers", a.opts.Concurrency),
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
	a.domainFilter.SetRoot(root)

	coordinator := NewCoordinator(
		a.processor,
		a.domainFilter,
		a.rateLimiter,
		a.opts.Depth,
		a.opts.Concurrency,
	)

	report.Pages = coordinator.Crawl(ctx, a.opts.URL)

	a.logger.Debug("analysis completed",
		zap.Int("pages_found", len(report.Pages)),
	)

	return report
}
