package extractor

import (
	"net/url"
	"strings"

	"code/internal/domain"
	"code/internal/usecase/analyzer/normalizer"

	"github.com/PuerkitoBio/goquery"
)

type LinkExtractor struct{}

func NewLinkExtractor() *LinkExtractor {
	return &LinkExtractor{}
}

func (e *LinkExtractor) Extract(pageURL string, doc *goquery.Document) []domain.Link {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}

	links := make([]domain.Link, 0, 32)
	seen := make(map[string]struct{}, 32)

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		attr, exists := s.Attr("href")
		if !exists {
			return
		}

		attr = strings.TrimSpace(attr)
		if attr == "" {
			return
		}

		resolved, ok := resolveURL(base, attr)
		if !ok {
			return
		}

		resolved = normalizer.NormalizeURL(resolved)

		if _, exists := seen[resolved]; exists {
			return
		}
		seen[resolved] = struct{}{}
		links = append(links, domain.Link{URL: resolved})
	})

	return links
}
