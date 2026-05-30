package extractor

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type AssetExtractor struct{}

func NewAssetExtractor() *AssetExtractor {
	return &AssetExtractor{}
}

func (e *AssetExtractor) Extract(pageURL string, doc *goquery.Document) []string {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}

	assets := make([]string, 0, 16)
	seen := make(map[string]struct{}, 16)

	doc.Find("img[src], script[src], link[rel='stylesheet']").Each(func(_ int, s *goquery.Selection) {
		tagName := goquery.NodeName(s)
		var attrVal string

		switch tagName {
		case "img", "script":
			attrVal, _ = s.Attr("src")
		case "link":
			attrVal, _ = s.Attr("href")
		}

		attrVal = strings.TrimSpace(attrVal)
		if attrVal == "" {
			return
		}

		resolved, ok := resolveURL(base, attrVal)
		if !ok {
			return
		}

		if _, exists := seen[resolved]; exists {
			return
		}
		seen[resolved] = struct{}{}
		assets = append(assets, resolved)
	})

	return assets
}
