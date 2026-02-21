package search

import (
  "sync/atomic"
  "testing"
  "time"

  "npan/internal/models"
)

// mockSearcher implements the Searcher interface for testing.
// It counts Query calls via an atomic counter and returns a preset result.
type mockSearcher struct {
  queryCalls atomic.Int64
  result     QueryResult
}

func (m *mockSearcher) Query(params models.LocalSearchParams) (QueryResult, error) {
  m.queryCalls.Add(1)
  return m.result, nil
}

func (m *mockSearcher) Ping() error {
  return nil
}

func newMockSearcher(result QueryResult) *mockSearcher {
  return &mockSearcher{result: result}
}

func TestCachedQueryService_CacheHit(t *testing.T) {
  mock := newMockSearcher(QueryResult{
    Items: []models.IndexDocument{{DocID: "doc-1", Name: "test.pdf"}},
    Total: 1,
  })

  cached := NewCachedQueryService(mock, 10, 5*time.Second, nil)

  params := models.LocalSearchParams{Query: "test", Page: 1, PageSize: 20}

  result1, err := cached.Query(params)
  if err != nil {
    t.Fatalf("first Query returned error: %v", err)
  }

  result2, err := cached.Query(params)
  if err != nil {
    t.Fatalf("second Query returned error: %v", err)
  }

  if result1.Total != result2.Total {
    t.Errorf("expected same Total, got %d and %d", result1.Total, result2.Total)
  }

  if calls := mock.queryCalls.Load(); calls != 1 {
    t.Errorf("expected mock to be called 1 time, got %d", calls)
  }
}

func TestCachedQueryService_CacheExpiry(t *testing.T) {
  mock := newMockSearcher(QueryResult{
    Items: []models.IndexDocument{{DocID: "doc-1", Name: "test.pdf"}},
    Total: 1,
  })

  cached := NewCachedQueryService(mock, 10, 50*time.Millisecond, nil)

  params := models.LocalSearchParams{Query: "test", Page: 1, PageSize: 20}

  _, err := cached.Query(params)
  if err != nil {
    t.Fatalf("first Query returned error: %v", err)
  }

  time.Sleep(100 * time.Millisecond)

  _, err = cached.Query(params)
  if err != nil {
    t.Fatalf("second Query returned error: %v", err)
  }

  if calls := mock.queryCalls.Load(); calls != 2 {
    t.Errorf("expected mock to be called 2 times after TTL expiry, got %d", calls)
  }
}

func TestCachedQueryService_DifferentParams(t *testing.T) {
  mock := newMockSearcher(QueryResult{
    Items: []models.IndexDocument{{DocID: "doc-1", Name: "test.pdf"}},
    Total: 1,
  })

  cached := NewCachedQueryService(mock, 10, 5*time.Second, nil)

  params1 := models.LocalSearchParams{Query: "alpha", Page: 1, PageSize: 20}
  params2 := models.LocalSearchParams{Query: "beta", Page: 1, PageSize: 20}

  _, err := cached.Query(params1)
  if err != nil {
    t.Fatalf("first Query returned error: %v", err)
  }

  _, err = cached.Query(params2)
  if err != nil {
    t.Fatalf("second Query returned error: %v", err)
  }

  if calls := mock.queryCalls.Load(); calls != 2 {
    t.Errorf("expected mock to be called 2 times for different params, got %d", calls)
  }
}

func TestCachedQueryService_LRUEviction(t *testing.T) {
  mock := newMockSearcher(QueryResult{
    Items: []models.IndexDocument{{DocID: "doc-1", Name: "test.pdf"}},
    Total: 1,
  })

  cached := NewCachedQueryService(mock, 2, 5*time.Second, nil)

  paramsA := models.LocalSearchParams{Query: "aaa", Page: 1, PageSize: 20}
  paramsB := models.LocalSearchParams{Query: "bbb", Page: 1, PageSize: 20}
  paramsC := models.LocalSearchParams{Query: "ccc", Page: 1, PageSize: 20}

  // Fill cache with A and B
  _, _ = cached.Query(paramsA)
  _, _ = cached.Query(paramsB)

  // Insert C, which should evict A (capacity=2)
  _, _ = cached.Query(paramsC)

  // At this point mock has been called 3 times (A, B, C)
  if calls := mock.queryCalls.Load(); calls != 3 {
    t.Fatalf("expected 3 calls after initial queries, got %d", calls)
  }

  // Touch B to make it most-recently-used, so C becomes oldest
  _, err := cached.Query(paramsB)
  if err != nil {
    t.Fatalf("Query for B (touch) returned error: %v", err)
  }
  if calls := mock.queryCalls.Load(); calls != 3 {
    t.Errorf("expected B to still be cached (3 calls), got %d", calls)
  }

  // Query A again — cache miss since A was evicted; adding A evicts C (oldest)
  _, err = cached.Query(paramsA)
  if err != nil {
    t.Fatalf("Query for evicted key returned error: %v", err)
  }

  if calls := mock.queryCalls.Load(); calls != 4 {
    t.Errorf("expected 4 calls after querying evicted key, got %d", calls)
  }

  // Query B — should still be cached because we touched it before inserting A
  _, err = cached.Query(paramsB)
  if err != nil {
    t.Fatalf("Query for cached key returned error: %v", err)
  }

  if calls := mock.queryCalls.Load(); calls != 4 {
    t.Errorf("expected B to still be cached (4 calls), got %d", calls)
  }
}
