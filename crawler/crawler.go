package crawler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"code/internal/infrastructure/httpclient"
	"code/internal/presenter"
	"code/internal/usecase/analyzer"

	"go.uber.org/zap"
)

var ErrHTTPClientRequired = errors.New("http client is required")

type Options struct {
	URL         string
	Depth       int
	Retries     int
	Delay       time.Duration
	Timeout     time.Duration
	RPS         float64
	UserAgent   string
	Concurrency int
	IndentJSON  bool
	HTTPClient  *http.Client
	Logger      *zap.Logger
}

func Analyze(ctx context.Context, opts Options) ([]byte, error) {
	if opts.HTTPClient == nil {
		return nil, ErrHTTPClientRequired
	}

	usecaseOpts := analyzer.Options{
		URL:         opts.URL,
		Depth:       opts.Depth,
		Retries:     opts.Retries,
		Delay:       opts.Delay,
		Timeout:     opts.Timeout,
		RPS:         opts.RPS,
		UserAgent:   opts.UserAgent,
		Concurrency: opts.Concurrency,
	}

	fetcher := httpclient.New(opts.HTTPClient, opts.UserAgent)
	defaultAnalyzer := analyzer.NewDefaultAnalyzer(opts.Logger, fetcher, usecaseOpts)
	report := defaultAnalyzer.Analyze(ctx)

	jsonPresenter := presenter.NewJSONPresenter(opts.IndentJSON)
	return jsonPresenter.Present(report)
}
