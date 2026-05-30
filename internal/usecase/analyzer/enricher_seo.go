package analyzer

import (
	"context"
	htmlpkg "html"
	"strings"

	"code/internal/domain"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type SEOEnricher struct {
	logger *zap.Logger
}

func NewSEOEnricher(logger *zap.Logger) *SEOEnricher {
	return &SEOEnricher{logger: logger}
}

func (e *SEOEnricher) Enrich(_ context.Context, page *domain.Page, doc *goquery.Document) {
	result := domain.SEOResult{}

	if titleSelection := doc.Find("title").First(); titleSelection.Length() > 0 {
		result.HasTitle = true
		result.Title = cleanText(titleSelection.Text())
	}

	doc.Find("meta[name]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		name, _ := s.Attr("name")
		if !strings.EqualFold(strings.TrimSpace(name), "description") {
			return true
		}

		content, _ := s.Attr("content")
		result.HasDescription = true
		result.Description = cleanText(content)
		return false
	})

	result.HasH1 = doc.Find("h1").Length() > 0

	page.SEO = &result
}

func cleanText(value string) string {
	unescaped := htmlpkg.UnescapeString(value)
	fields := strings.Fields(unescaped)
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, " ")
}
