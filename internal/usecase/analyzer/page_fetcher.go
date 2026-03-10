package analyzer

import (
	"context"

	"code/internal/domain"

	"go.uber.org/zap"
)

type PageFetchResult struct {
	Page        domain.Page
	Body        []byte
	IsHTML      bool
	StatusCode  int
	ContentType string
}

type PageFetcher interface {
	Fetch(ctx context.Context, url string, depth int) PageFetchResult
}

type pageFetcher struct {
	logger            *zap.Logger
	fetcher           Fetcher
	contentTypeFilter ContentTypeFilter
}

func NewPageFetcher(logger *zap.Logger, fetcher Fetcher, contentTypeFilter ContentTypeFilter) PageFetcher {
	return &pageFetcher{
		logger:            logger,
		fetcher:           fetcher,
		contentTypeFilter: contentTypeFilter,
	}
}

func (pf *pageFetcher) Fetch(ctx context.Context, pageURL string, depth int) PageFetchResult {
	pf.logger.Debug("fetching page",
		zap.String("url", pageURL),
		zap.Int("depth", depth),
	)

	page := domain.Page{
		URL:   pageURL,
		Depth: depth,
	}

	result, err := pf.fetcher.Fetch(ctx, pageURL)
	if err != nil {
		pf.logger.Debug("fetch failed",
			zap.String("url", pageURL),
			zap.Error(err),
		)
		page.Err = err
		return PageFetchResult{Page: page, IsHTML: false}
	}

	page.StatusCode = result.StatusCode
	isHTML := pf.contentTypeFilter.IsHTML(result.ContentType) && result.StatusCode < 400

	pf.logger.Debug("page fetched",
		zap.String("url", pageURL),
		zap.Int("status", result.StatusCode),
		zap.Bool("is_html", isHTML),
	)

	return PageFetchResult{
		Page:        page,
		Body:        result.Body,
		IsHTML:      isHTML,
		StatusCode:  result.StatusCode,
		ContentType: result.ContentType,
	}
}
