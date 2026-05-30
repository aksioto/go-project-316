package httpclient

import (
	"context"
	"io"
	"net/http"

	"code/internal/domain"
)

type Client struct {
	httpClient *http.Client
	userAgent  string
}

func New(httpClient *http.Client, userAgent string) *Client {
	return &Client{
		httpClient: httpClient,
		userAgent:  userAgent,
	}
}

func (c *Client) Fetch(ctx context.Context, url string) (domain.FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.FetchResult{}, err
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.FetchResult{}, err
	}
	defer resp.Body.Close()

	const maxBodySize = 10 << 20 // 10 MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return domain.FetchResult{}, err
	}

	return domain.FetchResult{
		StatusCode:    resp.StatusCode,
		Body:          body,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
	}, nil
}
