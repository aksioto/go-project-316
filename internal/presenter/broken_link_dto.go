package presenter

type BrokenLinkDTO struct {
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
}
