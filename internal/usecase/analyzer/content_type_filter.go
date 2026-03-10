package analyzer

import (
	"mime"
	"strings"
)

//// ContentTypeFilter determines if a content type represents an HTML page.
//type ContentTypeFilter interface {
//	IsHTML(contentType string) bool
//}

const (
	mimeTextHTML = "text/html"
	mimeXHTML    = "application/xhtml+xml"
)

type ContentTypeFilter struct{}

// NewContentTypeFilter creates a new ContentTypeFilter for HTML detection.
func NewContentTypeFilter() ContentTypeFilter {
	return ContentTypeFilter{}
}

func (f ContentTypeFilter) IsHTML(contentType string) bool {
	value := strings.ToLower(strings.TrimSpace(contentType))
	if value == "" {
		return false
	}

	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType == mimeTextHTML || mediaType == mimeXHTML
}
