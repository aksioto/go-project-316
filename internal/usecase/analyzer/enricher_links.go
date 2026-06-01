package analyzer

import (
	"context"
	"net/http"

	"code/internal/domain"
	"code/internal/usecase/analyzer/fetcher"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type LinksEnricher struct {
	logger        *zap.Logger
	fetcher       fetcher.Fetcher
	rateLimiter   Waiter
	linkExtractor LinkExtractor
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
		if err := e.rateLimiter.Wait(ctx); err != nil {
			break
		}

		result, err := e.fetcher.FetchHead(ctx, link.URL)
		if err != nil {
			e.logger.Debug("broken link (error)",
				zap.String("url", link.URL),
				zap.Error(err),
			)
			broken = append(broken, domain.BrokenLink{URL: link.URL, Err: err})
			continue
		}
		if result.StatusCode >= http.StatusBadRequest {
			e.logger.Debug("broken link (status)",
				zap.String("url", link.URL),
				zap.Int("status", result.StatusCode),
			)
			broken = append(broken, domain.BrokenLink{URL: link.URL, StatusCode: result.StatusCode})
		}
	}

	if len(broken) > 0 {
		e.logger.Debug("found broken links", zap.Int("count", len(broken)))
	}

	page.BrokenLinks = broken
}
