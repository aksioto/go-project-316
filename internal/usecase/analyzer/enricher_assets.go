package analyzer

import (
	"context"
	"strconv"
	"sync"

	"code/internal/domain"
	"code/internal/usecase/analyzer/extractor"
	"code/internal/usecase/analyzer/fetcher"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type AssetsEnricher struct {
	logger         *zap.Logger
	fetcher        fetcher.Fetcher
	rateLimiter    Waiter
	assetExtractor AssetExtractor
	cache          map[string]domain.Asset
	mu             sync.RWMutex
}

func NewAssetsEnricher(
	logger *zap.Logger,
	f fetcher.Fetcher,
	rl Waiter,
	ae AssetExtractor,
) *AssetsEnricher {
	return &AssetsEnricher{
		logger:         logger,
		fetcher:        f,
		rateLimiter:    rl,
		assetExtractor: ae,
		cache:          make(map[string]domain.Asset),
	}
}

func (e *AssetsEnricher) Enrich(ctx context.Context, page *domain.Page, doc *goquery.Document) {
	assetURLs := e.assetExtractor.Extract(page.URL, doc)
	if len(assetURLs) == 0 {
		return
	}

	assets := make([]domain.Asset, 0, len(assetURLs))

	for _, assetURL := range assetURLs {
		if ctx.Err() != nil {
			break
		}

		asset := e.checkAsset(ctx, assetURL)
		assets = append(assets, asset)
	}

	page.Assets = assets
}

func (e *AssetsEnricher) checkAsset(ctx context.Context, assetURL string) domain.Asset {
	e.mu.RLock()
	cached, found := e.cache[assetURL]
	e.mu.RUnlock()

	if found {
		e.logger.Debug("asset cache hit", zap.String("url", assetURL))
		return cached
	}

	asset := e.fetchAsset(ctx, assetURL)

	e.mu.Lock()
	e.cache[assetURL] = asset
	e.mu.Unlock()

	return asset
}

func (e *AssetsEnricher) fetchAsset(ctx context.Context, assetURL string) domain.Asset {
	assetType := extractor.GetAssetType(assetURL)

	asset := domain.Asset{
		URL:  assetURL,
		Type: assetType,
	}

	if err := e.rateLimiter.Wait(ctx); err != nil {
		asset.Error = err.Error()
		return asset
	}

	e.logger.Debug("fetching asset", zap.String("url", assetURL))

	result, err := e.fetcher.Fetch(ctx, assetURL)
	if err != nil {
		e.logger.Debug("asset fetch failed",
			zap.String("url", assetURL),
			zap.Error(err),
		)
		asset.Error = err.Error()
		return asset
	}

	asset.StatusCode = result.StatusCode

	if result.StatusCode >= 400 {
		asset.Error = "HTTP " + strconv.Itoa(result.StatusCode)
		return asset
	}

	size := e.getSize(result)
	asset.SizeBytes = size

	e.logger.Debug("asset fetched",
		zap.String("url", assetURL),
		zap.Int("status", result.StatusCode),
		zap.Int64("size", size),
	)

	return asset
}

func (e *AssetsEnricher) getSize(result domain.FetchResult) int64 {
	if result.ContentLength > 0 {
		return result.ContentLength
	}

	if len(result.Body) > 0 {
		return int64(len(result.Body))
	}

	return 0
}
