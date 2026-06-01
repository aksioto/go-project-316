package fetcher

import (
	"context"

	"code/internal/domain"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) (domain.FetchResult, error)
	FetchHead(ctx context.Context, url string) (domain.FetchResult, error)
}
