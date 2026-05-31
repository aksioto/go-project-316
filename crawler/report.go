package crawler

import "code/internal/presenter"

// Report is the JSON report produced by Analyze.
type Report = presenter.ReportDTO

// Page is the analysis result for a single page.
type Page = presenter.PageDTO

// SEO holds basic SEO metrics for a page.
type SEO = presenter.SEODTO

// BrokenLink describes a link that could not be reached.
type BrokenLink = presenter.BrokenLinkDTO

// Asset describes a static resource (image, script, style).
type Asset = presenter.AssetDTO
