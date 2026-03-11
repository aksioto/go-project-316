package analyzer

import (
	"bytes"
	"net/url"
	"strings"

	"code/internal/domain"

	"golang.org/x/net/html"
)

type AssetExtractor struct{}

func NewAssetExtractor() *AssetExtractor {
	return &AssetExtractor{}
}

func (e *AssetExtractor) Extract(pageURL string, body []byte) []string {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	assets := make([]string, 0)
	seen := make(map[string]struct{})

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			assetURL, assetType := extractAssetInfo(n)
			if assetURL != "" && assetType != "" {
				resolved, ok := resolveAssetURL(base, assetURL)
				if ok {
					if _, exists := seen[resolved]; !exists {
						seen[resolved] = struct{}{}
						assets = append(assets, resolved)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return assets
}

func (e *AssetExtractor) GetAssetType(assetURL string) domain.AssetType {
	lower := strings.ToLower(assetURL)

	if strings.HasSuffix(lower, ".css") {
		return domain.AssetStyle
	}

	if strings.HasSuffix(lower, ".js") {
		return domain.AssetScript
	}

	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".bmp"}
	for _, ext := range imageExts {
		if strings.HasSuffix(lower, ext) {
			return domain.AssetImage
		}
	}

	return domain.AssetOther
}

func extractAssetInfo(n *html.Node) (string, domain.AssetType) {
	switch n.Data {
	case "img":
		src := getAttr(n, "src")
		if src != "" {
			return src, domain.AssetImage
		}
	case "script":
		src := getAttr(n, "src")
		if src != "" {
			return src, domain.AssetScript
		}
	case "link":
		rel := getAttr(n, "rel")
		href := getAttr(n, "href")
		if href != "" && rel == "stylesheet" {
			return href, domain.AssetStyle
		}
	}
	return "", ""
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func resolveAssetURL(base *url.URL, raw string) (string, bool) {
	if strings.TrimSpace(raw) == "" {
		return "", false
	}

	if strings.HasPrefix(raw, "data:") {
		return "", false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	if parsed.Scheme == "" {
		parsed = base.ResolveReference(parsed)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", false
	}

	if parsed.Host == "" {
		return "", false
	}

	return parsed.String(), true
}
