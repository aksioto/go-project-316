package analyzer

// NewDefaultAnalyzer wires default dependencies for Analyzer.
func NewDefaultAnalyzer(fetcher Fetcher, opts Options) *Analyzer {
	extractor := NewLinkExtractor()
	checker := NewBrokenLinkChecker(fetcher)
	seoAnalyzer := NewSEOAnalyzer()

	return NewAnalyzer(fetcher, extractor, checker, seoAnalyzer, opts)
}
