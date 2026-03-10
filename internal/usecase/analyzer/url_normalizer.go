package analyzer

import (
	"net/url"
	"path"
	"strings"
)

var defaultIndexFiles = []string{
	"index.html",
	"index.htm",
	"index.php",
	"default.html",
	"default.htm",
}

func NormalizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	}

	hadTrailingSlash := strings.HasSuffix(u.Path, "/") && u.Path != "/"
	u.Path = path.Clean(u.Path)
	u.Path = normalizeIndexPath(u.Path)

	if hadTrailingSlash && !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}

	return u.String()
}

func normalizeIndexPath(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	for _, idx := range defaultIndexFiles {
		if strings.HasSuffix(p, "/"+idx) {
			return strings.TrimSuffix(p, idx)
		}
	}

	return p
}
