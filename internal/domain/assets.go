package domain

// Asset is a static resource of the page.
type Asset struct {
	URL        string
	Type       AssetType
	StatusCode int
	SizeBytes  int64
	Error      string
}

type AssetType string

const (
	AssetImage  AssetType = "image"
	AssetScript AssetType = "script"
	AssetStyle  AssetType = "style"
	AssetOther  AssetType = "other"
)
