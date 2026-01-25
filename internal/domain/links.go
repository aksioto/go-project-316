package domain

// Link represents a resolved link extracted from a page.
type Link struct {
	URL string
}

// BrokenLink is a result of link validation.
type BrokenLink struct {
	URL        string
	Err        error
	StatusCode int
}
