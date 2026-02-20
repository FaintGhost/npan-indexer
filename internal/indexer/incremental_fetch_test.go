package indexer

import (
	"context"
	"errors"
	"testing"

	"npan/internal/models"
	"npan/internal/npan"
)

func TestFetchIncrementalChanges_DeduplicateAndSplit(t *testing.T) {
	t.Parallel()

	since := int64(100)
	until := int64(200)

	calls := 0
	items, err := FetchIncrementalChanges(context.Background(), IncrementalFetchOptions{
		Since: since,
		Until: until,
		Retry: models.RetryPolicyOptions{MaxRetries: 1},
		Fetch: func(_ context.Context, start *int64, end *int64, pageID int64) (map[string]any, error) {
			calls++
			if start == nil || end == nil || *start != since || *end != until {
				t.Fatalf("unexpected time window: start=%v end=%v", start, end)
			}

			if pageID == 0 {
				return map[string]any{
					"page_count": 2,
					"files": []any{
						map[string]any{
							"id":          1,
							"name":        "A-v1",
							"modified_at": 10,
							"parent":      map[string]any{"id": 7},
						},
					},
					"folders": []any{
						map[string]any{
							"id":         2,
							"name":       "FolderDeleted",
							"is_deleted": true,
							"parent":     map[string]any{"id": 0},
						},
					},
				}, nil
			}

			return map[string]any{
				"page_count": 2,
				"files": []any{
					map[string]any{
						"id":          1,
						"name":        "A-v2",
						"modified_at": 12,
						"parent":      map[string]any{"id": 8},
					},
					map[string]any{
						"id":       3,
						"name":     "FileDeleted",
						"in_trash": true,
					},
				},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 change items, got %d", len(items))
	}

	byID := map[string]IncrementalInputItem{}
	for _, item := range items {
		byID[item.Doc.DocID] = item
	}

	file1, ok := byID["file_1"]
	if !ok {
		t.Fatal("missing file_1")
	}
	if file1.Deleted {
		t.Fatal("file_1 should be upsert")
	}
	if file1.Doc.Name != "A-v2" {
		t.Fatalf("expected deduplicated latest name A-v2, got %s", file1.Doc.Name)
	}

	folder2, ok := byID["folder_2"]
	if !ok {
		t.Fatal("missing folder_2")
	}
	if !folder2.Deleted {
		t.Fatal("folder_2 should be delete")
	}

	file3, ok := byID["file_3"]
	if !ok {
		t.Fatal("missing file_3")
	}
	if !file3.Deleted {
		t.Fatal("file_3 should be delete")
	}
}

func TestFetchIncrementalChanges_RetryOnRetriableError(t *testing.T) {
	t.Parallel()

	attempts := 0
	_, err := FetchIncrementalChanges(context.Background(), IncrementalFetchOptions{
		Since: 0,
		Until: 100,
		Retry: models.RetryPolicyOptions{MaxRetries: 2, BaseDelayMS: 1, MaxDelayMS: 1},
		Fetch: func(_ context.Context, start *int64, end *int64, pageID int64) (map[string]any, error) {
			attempts++
			if attempts == 1 {
				return nil, &npan.StatusError{Status: 500, Message: "server error"}
			}
			return map[string]any{"page_count": 1}, nil
		},
	})
	if err != nil {
		t.Fatalf("expected retry success, got err: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestFetchIncrementalChanges_NoRetryOnNonRetriableError(t *testing.T) {
	t.Parallel()

	attempts := 0
	_, err := FetchIncrementalChanges(context.Background(), IncrementalFetchOptions{
		Since: 0,
		Until: 100,
		Retry: models.RetryPolicyOptions{MaxRetries: 3, BaseDelayMS: 1, MaxDelayMS: 1},
		Fetch: func(_ context.Context, start *int64, end *int64, pageID int64) (map[string]any, error) {
			attempts++
			return nil, errors.New("bad request")
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected no retry, attempts=1, got %d", attempts)
	}
}
