package analyzer

import (
	"context"

	"code/internal/usecase/analyzer/extractor"

	"go.uber.org/zap"
)

type PageProcessor struct {
	logger        *zap.Logger
	pageFetcher   PageFetcher
	linkExtractor LinkExtractor
	enricher      PageEnricher
}

func NewPageProcessor(
	logger *zap.Logger,
	pageFetcher PageFetcher,
	linkExtractor LinkExtractor,
	enricher PageEnricher,
) *PageProcessor {
	return &PageProcessor{
		logger:        logger,
		pageFetcher:   pageFetcher,
		linkExtractor: linkExtractor,
		enricher:      enricher,
	}
}

func (p *PageProcessor) ProcessPage(ctx context.Context, pageURL string, depth int) pageResult {
	result := p.pageFetcher.Fetch(ctx, pageURL, depth)
	page := result.Page

	isStartPage := depth == 0
	hasError := page.Err != nil || (page.StatusCode >= 400 && page.StatusCode < 600)

	if !result.IsHTML {
		return pageResult{page: page, shouldInclude: isStartPage && hasError, depth: depth}
	}

	doc, err := extractor.ParseHTML(result.Body)
	if err != nil {
		p.logger.Debug("failed to parse HTML", zap.Error(err))
		return pageResult{page: page, shouldInclude: isStartPage, depth: depth}
	}

	links := p.linkExtractor.Extract(pageURL, doc)
	p.enricher.Enrich(ctx, &page, doc)

	return pageResult{page: page, links: links, shouldInclude: true, depth: depth}
}
