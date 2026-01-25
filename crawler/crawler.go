package crawler

import (
	"code/internal/usecase/analyzer"
	"context"
	"fmt"
	"net/http"
	"time"

	"code/internal/infrastructure/httpclient"
	"code/internal/presenter"
)

type Options struct {
	URL         string
	Depth       int
	Retries     int
	Delay       time.Duration
	Timeout     time.Duration
	UserAgent   string
	Concurrency int
	IndentJSON  bool
	HTTPClient  *http.Client
}

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	if opts.HTTPClient == nil {
		return nil, fmt.Errorf("http client is required")
	}

	usecaseOpts := analyzer.Options{
		URL:         opts.URL,
		Depth:       opts.Depth,
		Retries:     opts.Retries,
		Delay:       opts.Delay,
		Timeout:     opts.Timeout,
		UserAgent:   opts.UserAgent,
		Concurrency: opts.Concurrency,
	}

	fetcher := httpclient.New(opts.HTTPClient, opts.UserAgent)
	defaultAnalyzer := analyzer.NewDefaultAnalyzer(fetcher, usecaseOpts)
	report := defaultAnalyzer.Analyze(ctx)

	jsonPresenter := presenter.NewJSONPresenter(opts.IndentJSON)
	return jsonPresenter.Present(report)
}
