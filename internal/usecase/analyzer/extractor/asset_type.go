package extractor

import (
	"strings"

	"code/internal/domain"
)

func GetAssetType(assetURL string) domain.AssetType {
	lower := strings.ToLower(assetURL)

	if strings.HasSuffix(lower, ".css") {
		return domain.AssetStyle
	}

	if strings.HasSuffix(lower, ".js") {
		return domain.AssetScript
	}

	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".bmp"}
	for _, ext := range imageExts {
		if strings.HasSuffix(lower, ext) {
			return domain.AssetImage
		}
	}

	return domain.AssetOther
}
