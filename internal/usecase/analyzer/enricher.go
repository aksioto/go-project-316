package analyzer

import (
	"context"

	"code/internal/domain"

	"github.com/PuerkitoBio/goquery"
)

type PageEnricher interface {
	Enrich(ctx context.Context, page *domain.Page, doc *goquery.Document)
}

type CompositeEnricher struct {
	enrichers []PageEnricher
}

func NewCompositeEnricher(enrichers ...PageEnricher) *CompositeEnricher {
	return &CompositeEnricher{enrichers: enrichers}
}

func (c *CompositeEnricher) Enrich(ctx context.Context, page *domain.Page, doc *goquery.Document) {
	for _, e := range c.enrichers {
		if ctx.Err() != nil {
			return
		}
		e.Enrich(ctx, page, doc)
	}
}
