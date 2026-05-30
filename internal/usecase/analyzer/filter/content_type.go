package filter

import (
	"mime"
	"strings"
)

type ContentTypeFilter interface {
	IsHTML(contentType string) bool
}

const (
	mimeTextHTML = "text/html"
	mimeXHTML    = "application/xhtml+xml"
)

type contentTypeFilter struct{}

func NewContentTypeFilter() ContentTypeFilter {
	return &contentTypeFilter{}
}

func (f *contentTypeFilter) IsHTML(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == mimeTextHTML || mediaType == mimeXHTML
}
