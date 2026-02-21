package metrics_test

import (
	"context"
	"errors"
	"testing"

	"npan/internal/metrics"
	"npan/internal/models"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type mockIndexOperator struct {
	upsertErr   error
	deleteErr   error
	searchDocs  []models.IndexDocument
	searchTotal int64
	searchErr   error
	docCount    int64
	docCountErr error
}

func (m *mockIndexOperator) EnsureSettings(ctx context.Context) error { return nil }
func (m *mockIndexOperator) Ping() error                              { return nil }

func (m *mockIndexOperator) UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error {
	return m.upsertErr
}

func (m *mockIndexOperator) DeleteDocuments(ctx context.Context, docIDs []string) error {
	return m.deleteErr
}

func (m *mockIndexOperator) Search(params models.LocalSearchParams) ([]models.IndexDocument, int64, error) {
	return m.searchDocs, m.searchTotal, m.searchErr
}

func (m *mockIndexOperator) DocumentCount(ctx context.Context) (int64, error) {
	return m.docCount, m.docCountErr
}

func TestInstrumentedMeiliIndex_UpsertSuccess(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSearchMetrics(reg)
	mock := &mockIndexOperator{}
	inst := metrics.NewInstrumentedMeiliIndex(mock, sm)

	docs := make([]models.IndexDocument, 10)
	err := inst.UpsertDocuments(context.Background(), docs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v := testutil.ToFloat64(sm.MeiliUpsertedDocsTotal); v != 10 {
		t.Errorf("upserted docs: got %f, want 10", v)
	}
	if v := testutil.ToFloat64(sm.MeiliErrorsTotal.WithLabelValues("upsert")); v != 0 {
		t.Errorf("upsert errors: got %f, want 0", v)
	}
}

func TestInstrumentedMeiliIndex_UpsertError(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSearchMetrics(reg)
	mock := &mockIndexOperator{upsertErr: errors.New("fail")}
	inst := metrics.NewInstrumentedMeiliIndex(mock, sm)

	err := inst.UpsertDocuments(context.Background(), make([]models.IndexDocument, 5))
	if err == nil {
		t.Fatal("expected error")
	}

	if v := testutil.ToFloat64(sm.MeiliErrorsTotal.WithLabelValues("upsert")); v != 1 {
		t.Errorf("upsert errors: got %f, want 1", v)
	}
	if v := testutil.ToFloat64(sm.MeiliUpsertedDocsTotal); v != 0 {
		t.Errorf("upserted docs on error: got %f, want 0", v)
	}
}

func TestInstrumentedMeiliIndex_SearchSuccess(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSearchMetrics(reg)
	mock := &mockIndexOperator{searchDocs: []models.IndexDocument{{DocID: "1"}}, searchTotal: 1}
	inst := metrics.NewInstrumentedMeiliIndex(mock, sm)

	docs, total, err := inst.Search(models.LocalSearchParams{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 || total != 1 {
		t.Errorf("search result: got %d docs, %d total", len(docs), total)
	}
}

func TestInstrumentedMeiliIndex_SearchError(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSearchMetrics(reg)
	mock := &mockIndexOperator{searchErr: errors.New("fail")}
	inst := metrics.NewInstrumentedMeiliIndex(mock, sm)

	_, _, err := inst.Search(models.LocalSearchParams{Query: "test"})
	if err == nil {
		t.Fatal("expected error")
	}

	if v := testutil.ToFloat64(sm.MeiliErrorsTotal.WithLabelValues("search")); v != 1 {
		t.Errorf("search errors: got %f, want 1", v)
	}
}

func TestInstrumentedMeiliIndex_DocumentCount(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSearchMetrics(reg)
	mock := &mockIndexOperator{docCount: 1000}
	inst := metrics.NewInstrumentedMeiliIndex(mock, sm)

	count, err := inst.DocumentCount(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1000 {
		t.Errorf("count: got %d, want 1000", count)
	}
	if v := testutil.ToFloat64(sm.MeiliDocumentsTotal); v != 1000 {
		t.Errorf("MeiliDocumentsTotal gauge: got %f, want 1000", v)
	}
}

func TestInstrumentedMeiliIndex_DeleteError(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSearchMetrics(reg)
	mock := &mockIndexOperator{deleteErr: errors.New("fail")}
	inst := metrics.NewInstrumentedMeiliIndex(mock, sm)

	err := inst.DeleteDocuments(context.Background(), []string{"1"})
	if err == nil {
		t.Fatal("expected error")
	}

	if v := testutil.ToFloat64(sm.MeiliErrorsTotal.WithLabelValues("delete")); v != 1 {
		t.Errorf("delete errors: got %f, want 1", v)
	}
}
