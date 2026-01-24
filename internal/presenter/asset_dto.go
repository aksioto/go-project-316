package presenter

type AssetDTO struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Size   int64  `json:"size"`
	Status int    `json:"status"`
}
