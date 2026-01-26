package domain

// SEOResult is a result of the page SEO analysis.
type SEOResult struct {
	HasTitle       bool
	Title          string
	HasDescription bool
	Description    string
	HasH1          bool
}
