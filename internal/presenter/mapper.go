package presenter

import (
	"net/http"
	"sort"
	"time"

	"code/internal/domain"
)

func MapReport(r domain.Report) ReportDTO {
	pages := make([]PageDTO, 0, len(r.Pages))

	for _, p := range r.Pages {
		pages = append(pages, mapPage(p))
	}

	sort.SliceStable(pages, func(i, j int) bool {
		if pages[i].Depth != pages[j].Depth {
			return pages[i].Depth < pages[j].Depth
		}
		return pages[i].URL < pages[j].URL
	})

	return ReportDTO{
		RootURL:     r.RootURL,
		Depth:       r.MaxDepth,
		GeneratedAt: r.GeneratedAt,
		Pages:       pages,
	}
}

func mapPage(p domain.Page) PageDTO {
	dto := PageDTO{
		URL:        p.URL,
		Depth:      p.Depth,
		HTTPStatus: p.StatusCode,
		Status:     pageStatus(p.StatusCode, p.Err),
	}

	if p.Err != nil {
		dto.Error = p.Err.Error()
	}

	if p.SEO != nil {
		dto.SEO = mapSEO(*p.SEO)
	} else {
		dto.SEO = &SEODTO{}
	}

	if p.Err == nil {
		dto.BrokenLinks = make([]BrokenLinkDTO, 0)
		dto.Assets = make([]AssetDTO, 0)

		if len(p.Assets) > 0 {
			dto.Assets = mapAssets(p.Assets)
		}

		if len(p.BrokenLinks) > 0 {
			dto.BrokenLinks = mapBrokenLinks(p.BrokenLinks)
		}
	}

	if !p.DiscoveredAt.IsZero() {
		dto.DiscoveredAt = p.DiscoveredAt.Format(time.RFC3339)
	}

	return dto
}

func pageStatus(statusCode int, err error) string {
	if err != nil {
		return "error"
	}
	if statusCode >= http.StatusBadRequest {
		return "error"
	}
	return "ok"
}

func mapSEO(seo domain.SEOResult) *SEODTO {
	return &SEODTO{
		HasTitle:       seo.HasTitle,
		Title:          seo.Title,
		HasDescription: seo.HasDescription,
		Description:    seo.Description,
		HasH1:          seo.HasH1,
	}
}

func mapAssets(assets []domain.Asset) []AssetDTO {
	result := make([]AssetDTO, 0, len(assets))
	for _, a := range assets {
		result = append(result, AssetDTO{
			URL:        a.URL,
			Type:       string(a.Type),
			StatusCode: a.StatusCode,
			SizeBytes:  a.SizeBytes,
			Error:      a.Error,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type < result[j].Type
		}
		return result[i].URL < result[j].URL
	})

	return result
}

func mapBrokenLinks(links []domain.BrokenLink) []BrokenLinkDTO {
	result := make([]BrokenLinkDTO, 0, len(links))
	for _, l := range links {
		dto := BrokenLinkDTO{
			URL:        l.URL,
			StatusCode: l.StatusCode,
		}
		if l.Err != nil {
			dto.Error = l.Err.Error()
		}
		result = append(result, dto)
	}
	return result
}
