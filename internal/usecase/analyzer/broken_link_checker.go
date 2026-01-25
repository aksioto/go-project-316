package analyzer

import (
	"context"
	"net/http"

	"code/internal/domain"
)

type brokenLinkChecker struct {
	fetcher Fetcher
}

func NewBrokenLinkChecker(fetcher Fetcher) *brokenLinkChecker {
	return &brokenLinkChecker{
		fetcher: fetcher,
	}
}

func (c *brokenLinkChecker) Check(ctx context.Context, links []domain.Link) []domain.BrokenLink {
	if len(links) == 0 {
		return nil
	}

	broken := make([]domain.BrokenLink, 0)
	for _, link := range links {
		result, err := c.fetcher.Fetch(ctx, link.URL)
		if err != nil {
			broken = append(broken, domain.BrokenLink{URL: link.URL, Err: err})
			continue
		}
		if result.StatusCode >= http.StatusBadRequest {
			broken = append(broken, domain.BrokenLink{URL: link.URL, StatusCode: result.StatusCode})
		}
	}

	return broken
}
