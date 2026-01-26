package testutils

import "encoding/json"

type Report struct {
	RootURL string `json:"root_url"`
	Depth   int    `json:"depth"`
	Pages   []Page `json:"pages"`
}

type Page struct {
	URL         string       `json:"url"`
	HTTPStatus  int          `json:"http_status"`
	Status      string       `json:"status"`
	Error       string       `json:"error"`
	SEO         *SEO         `json:"seo,omitempty"`
	BrokenLinks []BrokenLink `json:"broken_links,omitempty"`
}

type SEO struct {
	HasTitle       bool   `json:"has_title"`
	Title          string `json:"title"`
	HasDescription bool   `json:"has_description"`
	Description    string `json:"description"`
	HasH1          bool   `json:"has_h1"`
}

type BrokenLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

func ParseReport(data []byte) (*Report, error) {
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}
