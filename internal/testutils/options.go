package testutils

import (
	"net/http"
	"time"

	"code/crawler"
)

const defaultTimeout = 5 * time.Second

func NewCrawlerOptions(url string, depth int, client *http.Client) crawler.Options {
	return crawler.Options{
		URL:        url,
		Depth:      depth,
		Timeout:    defaultTimeout,
		HTTPClient: client,
		IndentJSON: false,
	}
}
