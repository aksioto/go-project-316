package extractor

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
)

func ParseHTML(body []byte) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(bytes.NewReader(body))
}
