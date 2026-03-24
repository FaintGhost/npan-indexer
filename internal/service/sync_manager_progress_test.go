package service

import (
  "path/filepath"
  "testing"

  "npan/internal/models"
  "npan/internal/storage"
)

// TestGetProgress_RunningButStoreEmpty verifies that when a sync goroutine is
// active (m.running == true) but the progress store has no file yet (race
// window between Start() and the first progressStore.Save()), GetProgress
// returns a non-nil result with status "running".
func TestGetProgress_RunningButStoreEmpty(t *testing.T) {
  t.Parallel()

  store := storage.NewJSONProgressStore(filepath.Join(t.TempDir(), "progress.json"))
  m := &SyncManager{progressStore: store}

  // Simulate an active sync goroutine.
  m.mu.Lock()
  m.running = true
  m.mu.Unlock()

  progress, err := m.GetProgress()
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if progress == nil {
    t.Fatalf("expected non-nil progress when goroutine is running, got nil")
  }
  if progress.Status != "running" {
    t.Fatalf("expected status %q, got %q", "running", progress.Status)
  }
}

// TestGetProgress_RunningButStoreDone verifies that when a sync goroutine is
// active but the progress store still contains a stale record with
// status="done" (e.g. from a previous run), GetProgress returns status
// "running" to reflect the actual in-flight sync.
func TestGetProgress_RunningButStoreDone(t *testing.T) {
  t.Parallel()

  store := storage.NewJSONProgressStore(filepath.Join(t.TempDir(), "progress.json"))

  // Pre-populate the store with a completed sync record.
  if err := store.Save(&models.SyncProgressState{
    Status:     "done",
    StartedAt:  1000,
    UpdatedAt:  2000,
    RootProgress: map[string]*models.RootSyncProgress{},
  }); err != nil {
    t.Fatalf("failed to seed progress store: %v", err)
  }

  m := &SyncManager{progressStore: store}

  // Simulate an active sync goroutine.
  m.mu.Lock()
  m.running = true
  m.mu.Unlock()

  progress, err := m.GetProgress()
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if progress == nil {
    t.Fatalf("expected non-nil progress, got nil")
  }
  if progress.Status != "running" {
    t.Fatalf("expected status %q, got %q", "running", progress.Status)
  }
}

// TestGetProgress_RunningButStoreInterrupted verifies that when a sync
// goroutine is active but the progress store contains a record with
// status="interrupted" (e.g. written after a previous crash), GetProgress
// returns status "running" with an empty LastError to reflect the newly
// active sync.
func TestGetProgress_RunningButStoreInterrupted(t *testing.T) {
  t.Parallel()

  store := storage.NewJSONProgressStore(filepath.Join(t.TempDir(), "progress.json"))

  // Pre-populate the store with an interrupted sync record.
  if err := store.Save(&models.SyncProgressState{
    Status:       "interrupted",
    LastError:    "进程重启，同步中断",
    StartedAt:    1000,
    UpdatedAt:    2000,
    RootProgress: map[string]*models.RootSyncProgress{},
  }); err != nil {
    t.Fatalf("failed to seed progress store: %v", err)
  }

  m := &SyncManager{progressStore: store}

  // Simulate an active sync goroutine.
  m.mu.Lock()
  m.running = true
  m.mu.Unlock()

  progress, err := m.GetProgress()
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if progress == nil {
    t.Fatalf("expected non-nil progress, got nil")
  }
  if progress.Status != "running" {
    t.Fatalf("expected status %q, got %q", "running", progress.Status)
  }
  if progress.LastError != "" {
    t.Fatalf("expected empty LastError, got %q", progress.LastError)
  }
}

func TestGetProgress_MarksRunningRootsInterruptedAfterRestart(t *testing.T) {
  t.Parallel()

  store := storage.NewJSONProgressStore(filepath.Join(t.TempDir(), "progress.json"))

  currentFolderID := int64(200)
  currentPageID := int64(3)
  queueLength := int64(8)
  if err := store.Save(&models.SyncProgressState{
    Status:    "running",
    StartedAt: 1000,
    UpdatedAt: 2000,
    RootProgress: map[string]*models.RootSyncProgress{
      "100": {
        RootFolderID:     100,
        Status:           "running",
        CurrentFolderID:  &currentFolderID,
        CurrentPageID:    &currentPageID,
        QueueLength:      &queueLength,
        UpdatedAt:        2000,
      },
    },
  }); err != nil {
    t.Fatalf("failed to seed progress store: %v", err)
  }

  m := &SyncManager{progressStore: store}

  progress, err := m.GetProgress()
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if progress == nil {
    t.Fatalf("expected progress, got nil")
  }
  if progress.Status != "interrupted" {
    t.Fatalf("expected interrupted progress, got %q", progress.Status)
  }
  root := progress.RootProgress["100"]
  if root == nil {
    t.Fatalf("expected root progress")
  }
  if root.Status != "interrupted" {
    t.Fatalf("expected root interrupted, got %q", root.Status)
  }
  if root.CurrentFolderID != nil || root.CurrentPageID != nil || root.QueueLength != nil {
    t.Fatalf("expected stale running pointers to be cleared, got %+v", root)
  }
}
