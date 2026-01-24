package domain

// BrokenLink is a result of link validation.
type BrokenLink struct {
	URL string
	Err error
}
