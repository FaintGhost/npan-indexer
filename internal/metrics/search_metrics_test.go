package metrics_test

import (
	"testing"

	"npan/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewSearchMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewSearchMetrics(reg)

	// QueriesTotal
	m.QueriesTotal.WithLabelValues("hit").Inc()
	m.QueriesTotal.WithLabelValues("miss").Add(3)
	if v := testutil.ToFloat64(m.QueriesTotal.WithLabelValues("hit")); v != 1 {
		t.Errorf("QueriesTotal hit: got %f, want 1", v)
	}
	if v := testutil.ToFloat64(m.QueriesTotal.WithLabelValues("miss")); v != 3 {
		t.Errorf("QueriesTotal miss: got %f, want 3", v)
	}

	// CacheSize
	m.CacheSize.Set(42)
	if v := testutil.ToFloat64(m.CacheSize); v != 42 {
		t.Errorf("CacheSize: got %f, want 42", v)
	}

	// MeiliDurationSeconds
	m.MeiliDurationSeconds.WithLabelValues("search").Observe(0.05)

	// MeiliErrorsTotal
	m.MeiliErrorsTotal.WithLabelValues("search").Inc()
	if v := testutil.ToFloat64(m.MeiliErrorsTotal.WithLabelValues("search")); v != 1 {
		t.Errorf("MeiliErrorsTotal: got %f, want 1", v)
	}

	// MeiliDocumentsTotal
	m.MeiliDocumentsTotal.Set(1000)
	if v := testutil.ToFloat64(m.MeiliDocumentsTotal); v != 1000 {
		t.Errorf("MeiliDocumentsTotal: got %f, want 1000", v)
	}

	// MeiliUpsertedDocsTotal
	m.MeiliUpsertedDocsTotal.Add(50)
	if v := testutil.ToFloat64(m.MeiliUpsertedDocsTotal); v != 50 {
		t.Errorf("MeiliUpsertedDocsTotal: got %f, want 50", v)
	}
}
