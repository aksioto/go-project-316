package presenter

import (
	"encoding/json"
	"testing"
	"time"

	"code/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONPresenter_AllFieldsPresent(t *testing.T) {
	discoveredAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)
	generatedAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)

	report := domain.Report{
		RootURL:     "https://example.com",
		MaxDepth:    1,
		GeneratedAt: generatedAt,
		Pages: []domain.Page{
			{
				URL:          "https://example.com",
				Depth:        0,
				StatusCode:   200,
				DiscoveredAt: discoveredAt,
				SEO: &domain.SEOResult{
					HasTitle:       true,
					Title:          "Example title",
					HasDescription: true,
					Description:    "Example description",
					HasH1:          true,
				},
				BrokenLinks: []domain.BrokenLink{
					{
						URL:        "https://example.com/missing",
						StatusCode: 404,
						Err:        nil,
					},
				},
				Assets: []domain.Asset{
					{
						URL:        "https://example.com/static/logo.png",
						Type:       domain.AssetImage,
						StatusCode: 200,
						SizeBytes:  12345,
						Error:      "",
					},
				},
			},
		},
	}

	presenter := NewJSONPresenter(true)
	data, err := presenter.Present(report)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "https://example.com", result["root_url"])
	assert.Equal(t, float64(1), result["depth"])
	assert.Contains(t, result, "generated_at")

	pages := result["pages"].([]interface{})
	require.Len(t, pages, 1)

	page := pages[0].(map[string]interface{})
	assert.Equal(t, "https://example.com", page["url"])
	assert.Equal(t, float64(0), page["depth"])
	assert.Equal(t, float64(200), page["http_status"])
	assert.Equal(t, "ok", page["status"])
	assert.Contains(t, page, "error")
	assert.Contains(t, page, "seo")
	assert.Contains(t, page, "broken_links")
	assert.Contains(t, page, "assets")
	assert.Contains(t, page, "discovered_at")

	seo := page["seo"].(map[string]interface{})
	assert.Equal(t, true, seo["has_title"])
	assert.Equal(t, "Example title", seo["title"])
	assert.Equal(t, true, seo["has_description"])
	assert.Equal(t, "Example description", seo["description"])
	assert.Equal(t, true, seo["has_h1"])

	brokenLinks := page["broken_links"].([]interface{})
	require.Len(t, brokenLinks, 1)
	bl := brokenLinks[0].(map[string]interface{})
	assert.Equal(t, "https://example.com/missing", bl["url"])
	assert.Equal(t, float64(404), bl["status_code"])
	assert.Contains(t, bl, "error")

	assets := page["assets"].([]interface{})
	require.Len(t, assets, 1)
	asset := assets[0].(map[string]interface{})
	assert.Equal(t, "https://example.com/static/logo.png", asset["url"])
	assert.Equal(t, "image", asset["type"])
	assert.Equal(t, float64(200), asset["status_code"])
	assert.Equal(t, float64(12345), asset["size_bytes"])
	assert.Contains(t, asset, "error")
}

func TestJSONPresenter_EmptyFieldsPresent(t *testing.T) {
	generatedAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)
	discoveredAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)

	report := domain.Report{
		RootURL:     "https://example.com",
		MaxDepth:    1,
		GeneratedAt: generatedAt,
		Pages: []domain.Page{
			{
				URL:          "https://example.com",
				Depth:        0,
				StatusCode:   200,
				DiscoveredAt: discoveredAt,
				SEO:          nil,
				BrokenLinks:  nil,
				Assets:       nil,
			},
		},
	}

	presenter := NewJSONPresenter(false)
	data, err := presenter.Present(report)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	pages := result["pages"].([]interface{})
	page := pages[0].(map[string]interface{})

	assert.Contains(t, page, "error", "error field must always be present")
	assert.Contains(t, page, "broken_links", "broken_links field must always be present")
	assert.Contains(t, page, "assets", "assets field must always be present")
	assert.Contains(t, page, "discovered_at", "discovered_at field must always be present")

	brokenLinks := page["broken_links"].([]interface{})
	assert.Empty(t, brokenLinks, "broken_links should be empty array")

	assets := page["assets"].([]interface{})
	assert.Empty(t, assets, "assets should be empty array")
}

func TestJSONPresenter_IndentFormatting(t *testing.T) {
	generatedAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)
	discoveredAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)

	report := domain.Report{
		RootURL:     "https://example.com",
		MaxDepth:    1,
		GeneratedAt: generatedAt,
		Pages: []domain.Page{
			{
				URL:          "https://example.com",
				Depth:        0,
				StatusCode:   200,
				DiscoveredAt: discoveredAt,
			},
		},
	}

	indented := NewJSONPresenter(true)
	compact := NewJSONPresenter(false)

	indentedData, err := indented.Present(report)
	require.NoError(t, err)

	compactData, err := compact.Present(report)
	require.NoError(t, err)

	assert.Greater(t, len(indentedData), len(compactData), "indented JSON should be longer")
	assert.Contains(t, string(indentedData), "\n", "indented JSON should contain newlines")
	assert.NotContains(t, string(compactData), "\n", "compact JSON should not contain newlines")

	var indentedResult, compactResult map[string]interface{}
	err = json.Unmarshal(indentedData, &indentedResult)
	require.NoError(t, err)
	err = json.Unmarshal(compactData, &compactResult)
	require.NoError(t, err)

	assert.Equal(t, indentedResult, compactResult, "content should be identical regardless of formatting")
}

func TestJSONPresenter_ISO8601TimeFormat(t *testing.T) {
	generatedAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)
	discoveredAt := time.Date(2024, 6, 1, 12, 34, 57, 0, time.UTC)

	report := domain.Report{
		RootURL:     "https://example.com",
		MaxDepth:    1,
		GeneratedAt: generatedAt,
		Pages: []domain.Page{
			{
				URL:          "https://example.com",
				Depth:        0,
				StatusCode:   200,
				DiscoveredAt: discoveredAt,
			},
		},
	}

	presenter := NewJSONPresenter(false)
	data, err := presenter.Present(report)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	genAt := result["generated_at"].(string)
	_, err = time.Parse(time.RFC3339, genAt)
	assert.NoError(t, err, "generated_at should be in ISO8601/RFC3339 format")

	pages := result["pages"].([]interface{})
	page := pages[0].(map[string]interface{})
	discAt := page["discovered_at"].(string)
	_, err = time.Parse(time.RFC3339, discAt)
	assert.NoError(t, err, "discovered_at should be in ISO8601/RFC3339 format")
}

func TestJSONPresenter_MatchesReferenceStructure(t *testing.T) {
	generatedAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)
	discoveredAt := time.Date(2024, 6, 1, 12, 34, 56, 0, time.UTC)

	report := domain.Report{
		RootURL:     "https://example.com",
		MaxDepth:    1,
		GeneratedAt: generatedAt,
		Pages: []domain.Page{
			{
				URL:          "https://example.com",
				Depth:        0,
				StatusCode:   200,
				DiscoveredAt: discoveredAt,
				SEO: &domain.SEOResult{
					HasTitle:       true,
					Title:          "Example title",
					HasDescription: true,
					Description:    "Example description",
					HasH1:          true,
				},
				BrokenLinks: []domain.BrokenLink{
					{
						URL:        "https://example.com/missing",
						StatusCode: 404,
					},
				},
				Assets: []domain.Asset{
					{
						URL:        "https://example.com/static/logo.png",
						Type:       domain.AssetImage,
						StatusCode: 200,
						SizeBytes:  12345,
					},
				},
			},
		},
	}

	presenter := NewJSONPresenter(true)
	data, err := presenter.Present(report)
	require.NoError(t, err)

	var dto ReportDTO
	err = json.Unmarshal(data, &dto)
	require.NoError(t, err)

	assert.Equal(t, "https://example.com", dto.RootURL)
	assert.Equal(t, 1, dto.Depth)
	assert.False(t, dto.GeneratedAt.IsZero())

	require.Len(t, dto.Pages, 1)
	page := dto.Pages[0]

	assert.Equal(t, "https://example.com", page.URL)
	assert.Equal(t, 0, page.Depth)
	assert.Equal(t, 200, page.HTTPStatus)
	assert.Equal(t, "ok", page.Status)
	assert.NotNil(t, page.SEO)
	assert.NotNil(t, page.BrokenLinks)
	assert.NotNil(t, page.Assets)
	assert.NotEmpty(t, page.DiscoveredAt)

	assert.Equal(t, true, page.SEO.HasTitle)
	assert.Equal(t, "Example title", page.SEO.Title)

	require.Len(t, page.BrokenLinks, 1)
	assert.Equal(t, "https://example.com/missing", page.BrokenLinks[0].URL)
	assert.Equal(t, 404, page.BrokenLinks[0].StatusCode)

	require.Len(t, page.Assets, 1)
	assert.Equal(t, "https://example.com/static/logo.png", page.Assets[0].URL)
	assert.Equal(t, "image", page.Assets[0].Type)
	assert.Equal(t, 200, page.Assets[0].StatusCode)
	assert.Equal(t, int64(12345), page.Assets[0].SizeBytes)
}
