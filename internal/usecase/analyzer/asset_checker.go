package analyzer

import (
	"context"
	"strconv"
	"sync"

	"code/internal/domain"

	"go.uber.org/zap"
)

type AssetChecker struct {
	logger         *zap.Logger
	fetcher        Fetcher
	rateLimiter    *RateLimiter
	assetExtractor *AssetExtractor
	cache          map[string]domain.Asset
	mu             sync.RWMutex
}

func NewAssetChecker(
	logger *zap.Logger,
	fetcher Fetcher,
	rateLimiter *RateLimiter,
	assetExtractor *AssetExtractor,
) *AssetChecker {
	return &AssetChecker{
		logger:         logger,
		fetcher:        fetcher,
		rateLimiter:    rateLimiter,
		assetExtractor: assetExtractor,
		cache:          make(map[string]domain.Asset),
	}
}

func (c *AssetChecker) Check(ctx context.Context, pageURL string, body []byte) []domain.Asset {
	assetURLs := c.assetExtractor.Extract(pageURL, body)
	if len(assetURLs) == 0 {
		return nil
	}

	assets := make([]domain.Asset, 0, len(assetURLs))

	for _, assetURL := range assetURLs {
		if ctx.Err() != nil {
			break
		}

		asset := c.checkAsset(ctx, assetURL)
		assets = append(assets, asset)
	}

	return assets
}

func (c *AssetChecker) checkAsset(ctx context.Context, assetURL string) domain.Asset {
	c.mu.RLock()
	cached, found := c.cache[assetURL]
	c.mu.RUnlock()

	if found {
		c.logger.Debug("asset cache hit",
			zap.String("url", assetURL),
		)
		return cached
	}

	asset := c.fetchAsset(ctx, assetURL)

	c.mu.Lock()
	c.cache[assetURL] = asset
	c.mu.Unlock()

	return asset
}

func (c *AssetChecker) fetchAsset(ctx context.Context, assetURL string) domain.Asset {
	assetType := c.assetExtractor.GetAssetType(assetURL)

	asset := domain.Asset{
		URL:  assetURL,
		Type: assetType,
	}

	if err := c.rateLimiter.Wait(ctx); err != nil {
		asset.Error = err.Error()
		return asset
	}

	c.logger.Debug("fetching asset",
		zap.String("url", assetURL),
	)

	result, err := c.fetcher.Fetch(ctx, assetURL)
	if err != nil {
		c.logger.Debug("asset fetch failed",
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

	size := c.getSize(result)
	asset.SizeBytes = size

	c.logger.Debug("asset fetched",
		zap.String("url", assetURL),
		zap.Int("status", result.StatusCode),
		zap.Int64("size", size),
	)

	return asset
}

func (c *AssetChecker) getSize(result domain.FetchResult) int64 {
	if result.ContentLength > 0 {
		return result.ContentLength
	}

	if len(result.Body) > 0 {
		return int64(len(result.Body))
	}

	return 0
}

func (c *AssetChecker) CacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}
