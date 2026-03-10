package analyzer

import (
	"context"
	"net/http"

	"code/internal/domain"

	"go.uber.org/zap"
)

type BrokenLinkChecker struct {
	logger  *zap.Logger
	fetcher Fetcher
}

func NewBrokenLinkChecker(logger *zap.Logger, fetcher Fetcher) *BrokenLinkChecker {
	return &BrokenLinkChecker{
		logger:  logger,
		fetcher: fetcher,
	}
}

func (c *BrokenLinkChecker) Check(ctx context.Context, links []domain.Link) []domain.BrokenLink {
	if len(links) == 0 {
		return nil
	}

	c.logger.Debug("checking links", zap.Int("count", len(links)))

	broken := make([]domain.BrokenLink, 0, len(links))
	for _, link := range links {
		result, err := c.fetcher.Fetch(ctx, link.URL)
		if err != nil {
			c.logger.Debug("broken link (error)",
				zap.String("url", link.URL),
				zap.Error(err),
			)
			broken = append(broken, domain.BrokenLink{URL: link.URL, Err: err})
			continue
		}
		if result.StatusCode >= http.StatusBadRequest {
			c.logger.Debug("broken link (status)",
				zap.String("url", link.URL),
				zap.Int("status", result.StatusCode),
			)
			broken = append(broken, domain.BrokenLink{URL: link.URL, StatusCode: result.StatusCode})
		}
	}

	if len(broken) > 0 {
		c.logger.Debug("found broken links", zap.Int("count", len(broken)))
	}

	return broken
}
