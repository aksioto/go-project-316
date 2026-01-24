package domain

// Asset is a static resource of the page.
type Asset struct {
	URL        string
	Type       AssetType
	Size       int64
	StatusCode int
}

type AssetType string

const (
	AssetCSS   AssetType = "css"
	AssetJS    AssetType = "js"
	AssetImage AssetType = "image"
	AssetOther AssetType = "other"
)
