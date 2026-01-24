package presenter

type PageDTO struct {
	URL          string          `json:"url"`
	Depth        int             `json:"depth"`
	HTTPStatus   int             `json:"http_status"`
	Status       string          `json:"status"`
	Error        string          `json:"error,omitempty"`
	SEO          *SEODTO         `json:"seo,omitempty"`
	Assets       []AssetDTO      `json:"assets,omitempty"`
	BrokenLinks  []BrokenLinkDTO `json:"broken_links,omitempty"`
	DiscoveredAt *string         `json:"discovered_at,omitempty"`
}
