package presenter

import (
	"code/internal/domain"
	"time"
)

func MapReport(r domain.Report) ReportDTO {
	pages := make([]PageDTO, 0, len(r.Pages))

	for _, p := range r.Pages {
		pages = append(pages, mapPage(p))
	}

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
		Status:     statusFromError(p.Err),
	}

	if p.Err != nil {
		dto.Error = p.Err.Error()
	}

	if p.SEO != nil {
		dto.SEO = mapSEO(*p.SEO)
	}

	if len(p.Assets) > 0 {
		dto.Assets = mapAssets(p.Assets)
	}

	if len(p.BrokenLinks) > 0 {
		dto.BrokenLinks = mapBrokenLinks(p.BrokenLinks)
	}

	if !p.DiscoveredAt.IsZero() {
		t := p.DiscoveredAt.Format(time.RFC3339)
		dto.DiscoveredAt = &t
	}

	return dto
}

func statusFromError(err error) string {
	if err != nil {
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
			URL:    a.URL,
			Type:   string(a.Type),
			Size:   a.Size,
			Status: a.StatusCode,
		})
	}
	return result
}

func mapBrokenLinks(links []domain.BrokenLink) []BrokenLinkDTO {
	result := make([]BrokenLinkDTO, 0, len(links))
	for _, l := range links {
		dto := BrokenLinkDTO{
			URL: l.URL,
		}
		if l.StatusCode != 0 {
			dto.StatusCode = l.StatusCode
		}
		if l.Err != nil {
			dto.Error = l.Err.Error()
		}
		result = append(result, dto)
	}
	return result
}
