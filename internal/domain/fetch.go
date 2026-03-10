package domain

type FetchResult struct {
	StatusCode  int
	Body        []byte
	ContentType string
}
