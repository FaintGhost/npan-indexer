package service

import (
  "path/filepath"
  "testing"

  "npan/internal/models"
  "npan/internal/storage"
)

func TestNewSyncManager_AcceptsSQLiteStateStoresForProgressAndSyncState(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  stores, err := storage.NewSQLiteStateStores(storage.SQLiteStateStoresConfig{
    StateDBFile: filepath.Join(dir, "sync-state.sqlite"),
  })
  if err != nil {
    t.Fatalf("create sqlite stores failed: %v", err)
  }

  _ = NewSyncManager(SyncManagerArgs{
    ProgressStore:      stores.ProgressStore,
    SyncStateStore:     stores.SyncStateStore,
    CheckpointStores:   stores.CheckpointStoreFactory,
    MeiliHost:          "http://127.0.0.1:7700",
    MeiliIndex:         "test_items",
    CheckpointTemplate: filepath.Join(dir, "checkpoint.json"),
    RootWorkers:        1,
    ProgressEvery:      1,
    Retry: models.RetryPolicyOptions{
      MaxRetries:  1,
      BaseDelayMS: 1,
      MaxDelayMS:  1,
    },
    MaxConcurrent:   1,
    MinTimeMS:       0,
    IncrementalQuery: "*",
    WindowOverlapMS: 1000,
  })
}

func TestGetProgress_InterruptedStateIsPersistedBackToSQLite(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  stores, err := storage.NewSQLiteStateStores(storage.SQLiteStateStoresConfig{
    StateDBFile: filepath.Join(dir, "sync-state.sqlite"),
  })
  if err != nil {
    t.Fatalf("create sqlite stores failed: %v", err)
  }

  if err := stores.ProgressStore.Save(&models.SyncProgressState{
    Status:       "running",
    StartedAt:    1700000000000,
    UpdatedAt:    1700000000000,
    RootProgress: map[string]*models.RootSyncProgress{},
  }); err != nil {
    t.Fatalf("seed sqlite progress failed: %v", err)
  }

  mgr := NewSyncManager(SyncManagerArgs{
    ProgressStore:      stores.ProgressStore,
    SyncStateStore:     stores.SyncStateStore,
    CheckpointStores:   stores.CheckpointStoreFactory,
    MeiliHost:          "http://127.0.0.1:7700",
    MeiliIndex:         "test_items",
    CheckpointTemplate: filepath.Join(dir, "checkpoint.json"),
  })

  got, err := mgr.GetProgress()
  if err != nil {
    t.Fatalf("GetProgress failed: %v", err)
  }
  if got == nil {
    t.Fatal("expected non-nil progress")
  }
  if got.Status != "interrupted" {
    t.Fatalf("expected interrupted status, got %q", got.Status)
  }
  if got.LastError != "进程重启，同步中断" {
    t.Fatalf("expected interrupted error message, got %q", got.LastError)
  }

  persisted, err := stores.ProgressStore.Load()
  if err != nil {
    t.Fatalf("reload sqlite progress failed: %v", err)
  }
  if persisted == nil {
    t.Fatal("expected persisted progress in sqlite")
  }
  if persisted.Status != "interrupted" {
    t.Fatalf("expected sqlite persisted status interrupted, got %q", persisted.Status)
  }
  if persisted.LastError != "进程重启，同步中断" {
    t.Fatalf("expected sqlite persisted LastError, got %q", persisted.LastError)
  }
}
