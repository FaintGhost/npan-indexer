package metrics_test

import (
	"testing"

	"npan/internal/metrics"
	"npan/internal/models"
	"npan/internal/search"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type mockSearcher struct {
	result search.QueryResult
	err    error
	calls  int
}

func (m *mockSearcher) Query(params models.LocalSearchParams) (search.QueryResult, error) {
	m.calls++
	return m.result, m.err
}

func (m *mockSearcher) Ping() error { return nil }

type mockLenner struct {
	lengths []int
	idx     int
}

func (m *mockLenner) Len() int {
	if m.idx >= len(m.lengths) {
		return m.lengths[len(m.lengths)-1]
	}
	v := m.lengths[m.idx]
	m.idx++
	return v
}

func TestInstrumentedSearchService_CacheMiss(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSearchMetrics(reg)
	mock := &mockSearcher{result: search.QueryResult{Total: 5}}
	// Before: 0 entries, After: 1 entry → miss
	lenner := &mockLenner{lengths: []int{0, 1}}
	svc := metrics.NewInstrumentedSearchService(mock, lenner, sm)

	_, err := svc.Query(models.LocalSearchParams{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := testutil.ToFloat64(sm.QueriesTotal.WithLabelValues("miss")); v != 1 {
		t.Errorf("miss: got %f, want 1", v)
	}
	if v := testutil.ToFloat64(sm.QueriesTotal.WithLabelValues("hit")); v != 0 {
		t.Errorf("hit: got %f, want 0", v)
	}
	if v := testutil.ToFloat64(sm.CacheSize); v != 1 {
		t.Errorf("cache size: got %f, want 1", v)
	}
}

func TestInstrumentedSearchService_CacheHit(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSearchMetrics(reg)
	mock := &mockSearcher{result: search.QueryResult{Total: 5}}
	// Before: 1 entry, After: 1 entry (same size) → hit
	lenner := &mockLenner{lengths: []int{1, 1}}
	svc := metrics.NewInstrumentedSearchService(mock, lenner, sm)

	_, err := svc.Query(models.LocalSearchParams{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := testutil.ToFloat64(sm.QueriesTotal.WithLabelValues("hit")); v != 1 {
		t.Errorf("hit: got %f, want 1", v)
	}
	if v := testutil.ToFloat64(sm.QueriesTotal.WithLabelValues("miss")); v != 0 {
		t.Errorf("miss: got %f, want 0", v)
	}
}
