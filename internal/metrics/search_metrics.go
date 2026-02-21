package metrics

import "github.com/prometheus/client_golang/prometheus"

// SearchMetrics holds Prometheus metrics for search and Meilisearch operations.
type SearchMetrics struct {
	QueriesTotal           *prometheus.CounterVec
	CacheSize              prometheus.Gauge
	MeiliDurationSeconds   *prometheus.HistogramVec
	MeiliErrorsTotal       *prometheus.CounterVec
	MeiliDocumentsTotal    prometheus.Gauge
	MeiliUpsertedDocsTotal prometheus.Counter
}

// NewSearchMetrics creates and registers search metrics with the given registerer.
func NewSearchMetrics(reg prometheus.Registerer) *SearchMetrics {
	m := &SearchMetrics{
		QueriesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "npan_search_queries_total",
			Help: "Total number of search queries, partitioned by cache result.",
		}, []string{"result"}),
		CacheSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "npan_search_cache_size",
			Help: "Current number of entries in the search LRU cache.",
		}),
		MeiliDurationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "npan_meili_operation_duration_seconds",
			Help:    "Duration of Meilisearch operations in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		}, []string{"op"}),
		MeiliErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "npan_meili_operation_errors_total",
			Help: "Total number of Meilisearch operation errors by type.",
		}, []string{"op"}),
		MeiliDocumentsTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "npan_meili_documents_total",
			Help: "Total number of documents in the Meilisearch index.",
		}),
		MeiliUpsertedDocsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "npan_meili_upserted_docs_total",
			Help: "Total number of documents upserted to Meilisearch.",
		}),
	}
	reg.MustRegister(
		m.QueriesTotal,
		m.CacheSize,
		m.MeiliDurationSeconds,
		m.MeiliErrorsTotal,
		m.MeiliDocumentsTotal,
		m.MeiliUpsertedDocsTotal,
	)
	return m
}
