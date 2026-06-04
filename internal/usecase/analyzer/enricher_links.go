package analyzer

import (
	"context"
	"net/http"
	"sync"

	"code/internal/domain"
	"code/internal/usecase/analyzer/fetcher"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type linkCheckResult struct {
	broken     bool
	statusCode int
	err        error
}

type LinksEnricher struct {
	logger        *zap.Logger
	fetcher       fetcher.Fetcher
	rateLimiter   Waiter
	linkExtractor LinkExtractor
	cache         map[string]linkCheckResult
	mu            sync.RWMutex
}

func NewLinksEnricher(
	logger *zap.Logger,
	f fetcher.Fetcher,
	rl Waiter,
	le LinkExtractor,
) *LinksEnricher {
	return &LinksEnricher{
		logger:        logger,
		fetcher:       f,
		rateLimiter:   rl,
		linkExtractor: le,
		cache:         make(map[string]linkCheckResult),
	}
}

func (e *LinksEnricher) Enrich(ctx context.Context, page *domain.Page, doc *goquery.Document) {
	links := e.linkExtractor.Extract(page.URL, doc)
	if len(links) == 0 {
		return
	}

	e.logger.Debug("checking links", zap.Int("count", len(links)))

	broken := make([]domain.BrokenLink, 0)
	for _, link := range links {
		if ctx.Err() != nil {
			break
		}

		result := e.checkLink(ctx, link.URL)
		if result.broken {
			broken = append(broken, domain.BrokenLink{
				URL:        link.URL,
				StatusCode: result.statusCode,
				Err:        result.err,
			})
		}
	}

	if len(broken) > 0 {
		e.logger.Debug("found broken links", zap.Int("count", len(broken)))
	}

	page.BrokenLinks = broken
}

func (e *LinksEnricher) checkLink(ctx context.Context, url string) linkCheckResult {
	e.mu.RLock()
	cached, found := e.cache[url]
	e.mu.RUnlock()

	if found {
		e.logger.Debug("link cache hit", zap.String("url", url))
		return cached
	}

	result := e.fetchLink(ctx, url)

	e.mu.Lock()
	e.cache[url] = result
	e.mu.Unlock()

	return result
}

func (e *LinksEnricher) fetchLink(ctx context.Context, url string) linkCheckResult {
	if err := e.rateLimiter.Wait(ctx); err != nil {
		return linkCheckResult{broken: true, err: err}
	}

	result, err := e.fetcher.FetchHead(ctx, url)
	if err != nil {
		e.logger.Debug("broken link (error)",
			zap.String("url", url),
			zap.Error(err),
		)
		return linkCheckResult{broken: true, err: err}
	}

	if result.StatusCode >= http.StatusBadRequest {
		e.logger.Debug("broken link (status)",
			zap.String("url", url),
			zap.Int("status", result.StatusCode),
		)
		return linkCheckResult{broken: true, statusCode: result.StatusCode}
	}

	return linkCheckResult{broken: false, statusCode: result.StatusCode}
}
