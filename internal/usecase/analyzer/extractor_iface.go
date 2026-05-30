package analyzer

import (
	"code/internal/domain"

	"github.com/PuerkitoBio/goquery"
)

type LinkExtractor interface {
	Extract(pageURL string, doc *goquery.Document) []domain.Link
}

type AssetExtractor interface {
	Extract(pageURL string, doc *goquery.Document) []string
}
