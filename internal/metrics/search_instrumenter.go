package metrics

import (
	"npan/internal/models"
	"npan/internal/search"
)

// CacheLenner provides cache size information.
type CacheLenner interface {
	Len() int
}

// InstrumentedSearchService wraps search.Searcher with cache hit/miss metrics.
type InstrumentedSearchService struct {
	inner   search.Searcher
	lenner  CacheLenner
	metrics *SearchMetrics
}

// NewInstrumentedSearchService creates a new instrumented search service decorator.
func NewInstrumentedSearchService(inner search.Searcher, lenner CacheLenner, m *SearchMetrics) *InstrumentedSearchService {
	return &InstrumentedSearchService{inner: inner, lenner: lenner, metrics: m}
}

func (s *InstrumentedSearchService) Query(params models.LocalSearchParams) (search.QueryResult, error) {
	lenBefore := s.lenner.Len()
	result, err := s.inner.Query(params)
	if err != nil {
		return result, err
	}

	lenAfter := s.lenner.Len()
	if lenAfter > lenBefore {
		s.metrics.QueriesTotal.WithLabelValues("miss").Inc()
	} else {
		s.metrics.QueriesTotal.WithLabelValues("hit").Inc()
	}
	s.metrics.CacheSize.Set(float64(lenAfter))

	return result, nil
}

func (s *InstrumentedSearchService) Ping() error {
	return s.inner.Ping()
}
