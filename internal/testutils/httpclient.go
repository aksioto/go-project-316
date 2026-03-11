package testutils

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type RoundTripFunc func(*http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func NewErrorClient(err error) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, err
		}),
	}
}

func NewResponseClient(status int, body string) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			headers := make(http.Header)
			headers.Set("Content-Type", "text/html; charset=utf-8")
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     headers,
				Request:    req,
			}, nil
		}),
	}
}

func NewTimeoutClient() *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		}),
	}
}

type StubResponse struct {
	StatusCode  int
	Body        string
	Err         error
	ContentType string
}

func NewStubClientFunc(fn func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func NewStubClient(responses map[string]StubResponse) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			resp, ok := responses[req.URL.String()]
			if !ok {
				return nil, fmt.Errorf("unexpected url: %s", req.URL.String())
			}
			if resp.Err != nil {
				return nil, resp.Err
			}
			contentType := resp.ContentType
			if strings.TrimSpace(contentType) == "" {
				contentType = "text/html; charset=utf-8"
			}
			headers := make(http.Header)
			headers.Set("Content-Type", contentType)
			return &http.Response{
				StatusCode: resp.StatusCode,
				Body:       io.NopCloser(strings.NewReader(resp.Body)),
				Header:     headers,
				Request:    req,
			}, nil
		}),
	}
}
