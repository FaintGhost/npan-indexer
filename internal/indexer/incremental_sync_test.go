package indexer

import (
	"context"
	"errors"
	"testing"

	"npan/internal/models"
)

type fakeSyncStateStore struct {
	loaded *models.SyncState
	saved  *models.SyncState
}

func (s *fakeSyncStateStore) Load() (*models.SyncState, error) {
	return s.loaded, nil
}

func (s *fakeSyncStateStore) Save(state *models.SyncState) error {
	copied := *state
	s.saved = &copied
	return nil
}

func TestRunIncrementalSync_SaveCursorOnSuccess(t *testing.T) {
	t.Parallel()

	store := &fakeSyncStateStore{loaded: &models.SyncState{LastSyncTime: 100}}
	err := RunIncrementalSync(context.Background(), IncrementalDeps{
		FetchChanges: func(ctx context.Context, since int64) ([]IncrementalInputItem, error) {
			if since != 100 {
				t.Fatalf("expected since=100, got %d", since)
			}
			return []IncrementalInputItem{{
				Doc: models.IndexDocument{DocID: "file_1"},
			}}, nil
		},
		StateStore: store,
		UpsertDocuments: func(ctx context.Context, docs []models.IndexDocument) error {
			if len(docs) != 1 {
				t.Fatalf("expected 1 upsert, got %d", len(docs))
			}
			return nil
		},
		DeleteDocuments: func(ctx context.Context, docIDs []string) error {
			if len(docIDs) != 0 {
				t.Fatalf("expected no deletes, got %d", len(docIDs))
			}
			return nil
		},
		NowProvider: func() int64 { return 999 },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.saved == nil || store.saved.LastSyncTime != 999 {
		t.Fatalf("expected saved cursor 999, got %#v", store.saved)
	}
}

func TestRunIncrementalSync_DoNotSaveCursorOnUpsertError(t *testing.T) {
	t.Parallel()

	store := &fakeSyncStateStore{loaded: &models.SyncState{LastSyncTime: 100}}
	err := RunIncrementalSync(context.Background(), IncrementalDeps{
		FetchChanges: func(ctx context.Context, since int64) ([]IncrementalInputItem, error) {
			return []IncrementalInputItem{{Doc: models.IndexDocument{DocID: "file_1"}}}, nil
		},
		StateStore: store,
		UpsertDocuments: func(ctx context.Context, docs []models.IndexDocument) error {
			return errors.New("upsert failed")
		},
		DeleteDocuments: func(ctx context.Context, docIDs []string) error {
			return nil
		},
		NowProvider: func() int64 { return 999 },
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if store.saved != nil {
		t.Fatalf("cursor should not be saved on failure, got %#v", store.saved)
	}
}

func TestRunIncrementalSync_NormalizeMillisecondCursor(t *testing.T) {
	t.Parallel()

	store := &fakeSyncStateStore{loaded: &models.SyncState{LastSyncTime: 1_771_604_573_376}}
	err := RunIncrementalSync(context.Background(), IncrementalDeps{
		FetchChanges: func(ctx context.Context, since int64) ([]IncrementalInputItem, error) {
			if since != 1_771_604_573 {
				t.Fatalf("expected normalized since=1771604573, got %d", since)
			}
			return nil, nil
		},
		StateStore:      store,
		UpsertDocuments: func(ctx context.Context, docs []models.IndexDocument) error { return nil },
		DeleteDocuments: func(ctx context.Context, docIDs []string) error { return nil },
		NowProvider:     func() int64 { return 1_771_604_600_123 },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.saved == nil || store.saved.LastSyncTime != 1_771_604_600 {
		t.Fatalf("expected normalized saved cursor 1771604600, got %#v", store.saved)
	}
}
