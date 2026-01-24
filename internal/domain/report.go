package domain

import "time"

// Report is an aggregate of the crawl result.
type Report struct {
	RootURL     string
	MaxDepth    int
	GeneratedAt time.Time
	Pages       []Page
}

// Page is the analysis result for a single page.
type Page struct {
	URL          string
	Depth        int
	StatusCode   int
	Err          error
	SEO          *SEOResult
	Assets       []Asset
	BrokenLinks  []BrokenLink
	DiscoveredAt time.Time
}
