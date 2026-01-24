package crawler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"code/internal/infrastructure/httpclient"
	"code/internal/presenter"
	"code/internal/usecase"
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

	usecaseOpts := usecase.Options{
		URL:         opts.URL,
		Depth:       opts.Depth,
		Retries:     opts.Retries,
		Delay:       opts.Delay,
		Timeout:     opts.Timeout,
		UserAgent:   opts.UserAgent,
		Concurrency: opts.Concurrency,
	}

	fetcher := httpclient.New(opts.HTTPClient, opts.UserAgent)
	analyzer := usecase.NewAnalyzer(fetcher, usecaseOpts)
	report := analyzer.Analyze(ctx)

	jsonPresenter := presenter.NewJSONPresenter(opts.IndentJSON)
	return jsonPresenter.Present(report)
}
