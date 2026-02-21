package metrics

import (
	"context"
	"time"

	"npan/internal/models"
	"npan/internal/search"
)

// InstrumentedMeiliIndex wraps search.IndexOperator with Prometheus metrics.
type InstrumentedMeiliIndex struct {
	inner   search.IndexOperator
	metrics *SearchMetrics
}

// NewInstrumentedMeiliIndex creates a new instrumented decorator.
func NewInstrumentedMeiliIndex(inner search.IndexOperator, m *SearchMetrics) *InstrumentedMeiliIndex {
	return &InstrumentedMeiliIndex{inner: inner, metrics: m}
}

func (i *InstrumentedMeiliIndex) EnsureSettings(ctx context.Context) error {
	return i.inner.EnsureSettings(ctx)
}

func (i *InstrumentedMeiliIndex) UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error {
	start := time.Now()
	err := i.inner.UpsertDocuments(ctx, docs)
	i.metrics.MeiliDurationSeconds.WithLabelValues("upsert").Observe(time.Since(start).Seconds())
	if err != nil {
		i.metrics.MeiliErrorsTotal.WithLabelValues("upsert").Inc()
		return err
	}
	i.metrics.MeiliUpsertedDocsTotal.Add(float64(len(docs)))
	return nil
}

func (i *InstrumentedMeiliIndex) DeleteDocuments(ctx context.Context, docIDs []string) error {
	start := time.Now()
	err := i.inner.DeleteDocuments(ctx, docIDs)
	i.metrics.MeiliDurationSeconds.WithLabelValues("delete").Observe(time.Since(start).Seconds())
	if err != nil {
		i.metrics.MeiliErrorsTotal.WithLabelValues("delete").Inc()
	}
	return err
}

func (i *InstrumentedMeiliIndex) Search(params models.LocalSearchParams) ([]models.IndexDocument, int64, error) {
	start := time.Now()
	docs, total, err := i.inner.Search(params)
	i.metrics.MeiliDurationSeconds.WithLabelValues("search").Observe(time.Since(start).Seconds())
	if err != nil {
		i.metrics.MeiliErrorsTotal.WithLabelValues("search").Inc()
	}
	return docs, total, err
}

func (i *InstrumentedMeiliIndex) Ping() error {
	return i.inner.Ping()
}

func (i *InstrumentedMeiliIndex) DocumentCount(ctx context.Context) (int64, error) {
	start := time.Now()
	count, err := i.inner.DocumentCount(ctx)
	i.metrics.MeiliDurationSeconds.WithLabelValues("document_count").Observe(time.Since(start).Seconds())
	if err != nil {
		i.metrics.MeiliErrorsTotal.WithLabelValues("document_count").Inc()
		return 0, err
	}
	i.metrics.MeiliDocumentsTotal.Set(float64(count))
	return count, nil
}
