package presenter

type PageDTO struct {
	URL          string          `json:"url"`
	Depth        int             `json:"depth"`
	HTTPStatus   int             `json:"http_status"`
	Status       string          `json:"status"`
	Error        string          `json:"error"`
	SEO          *SEODTO         `json:"seo"`
	BrokenLinks  []BrokenLinkDTO `json:"broken_links"`
	Assets       []AssetDTO      `json:"assets"`
	DiscoveredAt string          `json:"discovered_at"`
}
