package analyzer

import (
	"bytes"
	htmlpkg "html"
	"strings"

	"code/internal/domain"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type SEOAnalyzer struct {
	logger *zap.Logger
}

func NewSEOAnalyzer(logger *zap.Logger) *SEOAnalyzer {
	return &SEOAnalyzer{logger: logger}
}

func (a *SEOAnalyzer) Analyze(body []byte) domain.SEOResult {
	result := domain.SEOResult{}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		a.logger.Debug("seo: failed to parse HTML", zap.Error(err))
		return result
	}

	if titleSelection := doc.Find("title").First(); titleSelection.Length() > 0 {
		result.HasTitle = true
		result.Title = cleanText(titleSelection.Text())
	}

	doc.Find("meta[name]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		name, _ := s.Attr("name")
		if !strings.EqualFold(strings.TrimSpace(name), "description") {
			return true
		}

		content, _ := s.Attr("content")
		result.HasDescription = true
		result.Description = cleanText(content)
		return false
	})

	result.HasH1 = doc.Find("h1").Length() > 0

	return result
}

func cleanText(value string) string {
	unescaped := htmlpkg.UnescapeString(value)
	fields := strings.Fields(unescaped)
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, " ")
}
