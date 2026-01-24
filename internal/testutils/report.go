package testutils

import "encoding/json"

type Report struct {
	RootURL string `json:"root_url"`
	Depth   int    `json:"depth"`
	Pages   []Page `json:"pages"`
}

type Page struct {
	URL        string `json:"url"`
	HTTPStatus int    `json:"http_status"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

func ParseReport(data []byte) (*Report, error) {
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}
