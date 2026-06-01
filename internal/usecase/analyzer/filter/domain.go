package filter

import (
	"net/url"
	"strings"
)

type DomainFilter interface {
	IsSameDomain(link string) bool
	SetRoot(root *url.URL)
}

type domainFilter struct {
	root *url.URL
}

func NewDomainFilter() DomainFilter {
	return &domainFilter{}
}

func (f *domainFilter) IsSameDomain(link string) bool {
	parsed, err := url.Parse(link)
	if err != nil {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	if f.root.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Host, f.root.Host)
}

func (f *domainFilter) SetRoot(root *url.URL) {
	f.root = root
}
