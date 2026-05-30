package extractor

import (
	"net/url"
	"strings"
)

func resolveURL(base *url.URL, raw string) (string, bool) {
	if strings.HasPrefix(raw, "data:") {
		return "", false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	if parsed.Scheme == "" {
		parsed = base.ResolveReference(parsed)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", false
	}

	if parsed.Host == "" {
		return "", false
	}

	return parsed.String(), true
}
