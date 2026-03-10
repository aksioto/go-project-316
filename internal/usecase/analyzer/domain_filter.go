package analyzer

import (
	"net/url"
	"strings"
)

//type DomainFilter interface {
//	IsSameDomain(rootURL *url.URL, link string) bool
//}

type DomainFilter struct{}

func NewDomainFilter() DomainFilter {
	return DomainFilter{}
}

func (f DomainFilter) IsSameDomain(root *url.URL, link string) bool {
	parsed, err := url.Parse(link)
	if err != nil {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	if root.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Host, root.Host)
}
