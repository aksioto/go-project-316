package testutils

import (
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
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
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
