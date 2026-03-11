package domain

// Asset is a static resource of the page.
type Asset struct {
	URL        string    `json:"url"`
	Type       AssetType `json:"type"`
	StatusCode int       `json:"status_code"`
	SizeBytes  int64     `json:"size_bytes"`
	Error      string    `json:"error"`
}

type AssetType string

const (
	AssetImage  AssetType = "image"
	AssetScript AssetType = "script"
	AssetStyle  AssetType = "style"
	AssetOther  AssetType = "other"
)
