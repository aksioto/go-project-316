package analyzer

import (
	"bytes"
	"net/url"
	"strings"

	"code/internal/domain"

	"golang.org/x/net/html"
)

type linkExtractor struct{}

func NewLinkExtractor() *linkExtractor {
	return &linkExtractor{}
}

func (e *linkExtractor) Extract(pageURL string, body []byte) []domain.Link {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	links := make([]domain.Link, 0)
	seen := make(map[string]struct{})
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			attr := linkAttribute(n)
			if attr != "" {
				resolved, ok := resolveLink(base, attr)
				if ok {
					if _, exists := seen[resolved]; !exists {
						seen[resolved] = struct{}{}
						links = append(links, domain.Link{URL: resolved})
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return links
}

func linkAttribute(n *html.Node) string {
	var key string
	switch n.Data {
	case "a", "link":
		key = "href"
	case "script", "img":
		key = "src"
	default:
		return ""
	}
	for _, attr := range n.Attr {
		if attr.Key == key {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func resolveLink(base *url.URL, raw string) (string, bool) {
	if strings.TrimSpace(raw) == "" {
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
