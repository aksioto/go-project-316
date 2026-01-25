package analyzer

// NewDefaultAnalyzer wires default dependencies for Analyzer.
func NewDefaultAnalyzer(fetcher Fetcher, opts Options) *Analyzer {
	extractor := NewLinkExtractor()
	checker := NewBrokenLinkChecker(fetcher)

	return NewAnalyzer(fetcher, extractor, checker, opts)
}
