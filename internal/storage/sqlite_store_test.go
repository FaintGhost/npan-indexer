package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"npan/internal/models"
)

func TestNewSQLiteStateStores_StateDBSeparatesNamespaces(t *testing.T) {
	dir := t.TempDir()
	stores, err := NewSQLiteStateStores(SQLiteStateStoresConfig{
		StateDBFile: filepath.Join(dir, "sync-state.sqlite"),
	})
	if err != nil {
		t.Fatalf("create sqlite stores failed: %v", err)
	}

	progress := sampleSyncProgressState(101)
	syncState := sampleSyncState(1_710_000_000_000)
	checkpointKeyA := filepath.Join(dir, "checkpoint-a.json")
	checkpointKeyB := filepath.Join(dir, "checkpoint-b.json")
	checkpointA := sampleCheckpoint([]int64{1, 2, 3}, 11, 21)
	checkpointB := sampleCheckpoint([]int64{7, 8, 9}, 17, 27)

	if err := stores.ProgressStore.Save(progress); err != nil {
		t.Fatalf("save progress failed: %v", err)
	}
	if err := stores.SyncStateStore.Save(syncState); err != nil {
		t.Fatalf("save sync state failed: %v", err)
	}
	if err := stores.CheckpointStoreFactory.ForKey(checkpointKeyA).Save(checkpointA); err != nil {
		t.Fatalf("save checkpoint A failed: %v", err)
	}
	if err := stores.CheckpointStoreFactory.ForKey(checkpointKeyB).Save(checkpointB); err != nil {
		t.Fatalf("save checkpoint B failed: %v", err)
	}

	loadedProgress, err := stores.ProgressStore.Load()
	if err != nil {
		t.Fatalf("load progress failed: %v", err)
	}
	if !reflect.DeepEqual(loadedProgress, progress) {
		t.Fatalf("unexpected progress: %#v", loadedProgress)
	}

	loadedSyncState, err := stores.SyncStateStore.Load()
	if err != nil {
		t.Fatalf("load sync state failed: %v", err)
	}
	if !reflect.DeepEqual(loadedSyncState, syncState) {
		t.Fatalf("unexpected sync state: %#v", loadedSyncState)
	}

	loadedCheckpointA, err := stores.CheckpointStoreFactory.ForKey(checkpointKeyA).Load()
	if err != nil {
		t.Fatalf("load checkpoint A failed: %v", err)
	}
	if !reflect.DeepEqual(loadedCheckpointA, checkpointA) {
		t.Fatalf("unexpected checkpoint A: %#v", loadedCheckpointA)
	}

	loadedCheckpointB, err := stores.CheckpointStoreFactory.ForKey(checkpointKeyB).Load()
	if err != nil {
		t.Fatalf("load checkpoint B failed: %v", err)
	}
	if !reflect.DeepEqual(loadedCheckpointB, checkpointB) {
		t.Fatalf("unexpected checkpoint B: %#v", loadedCheckpointB)
	}

	if err := stores.CheckpointStoreFactory.ForKey(checkpointKeyA).Clear(); err != nil {
		t.Fatalf("clear checkpoint A failed: %v", err)
	}

	loadedCheckpointA, err = stores.CheckpointStoreFactory.ForKey(checkpointKeyA).Load()
	if err != nil {
		t.Fatalf("reload checkpoint A failed: %v", err)
	}
	if loadedCheckpointA != nil {
		t.Fatalf("expected checkpoint A to be cleared, got %#v", loadedCheckpointA)
	}

	loadedCheckpointB, err = stores.CheckpointStoreFactory.ForKey(checkpointKeyB).Load()
	if err != nil {
		t.Fatalf("reload checkpoint B failed: %v", err)
	}
	if !reflect.DeepEqual(loadedCheckpointB, checkpointB) {
		t.Fatalf("checkpoint B should remain intact: %#v", loadedCheckpointB)
	}

	loadedProgress, err = stores.ProgressStore.Load()
	if err != nil {
		t.Fatalf("reload progress failed: %v", err)
	}
	if !reflect.DeepEqual(loadedProgress, progress) {
		t.Fatalf("progress should remain intact: %#v", loadedProgress)
	}

	loadedSyncState, err = stores.SyncStateStore.Load()
	if err != nil {
		t.Fatalf("reload sync state failed: %v", err)
	}
	if !reflect.DeepEqual(loadedSyncState, syncState) {
		t.Fatalf("sync state should remain intact: %#v", loadedSyncState)
	}
}

func TestSQLiteProgressStore_LegacyImport_PersistsIntoStateDB(t *testing.T) {
	dir := t.TempDir()
	legacyFile := filepath.Join(dir, "legacy-progress.json")
	want := sampleSyncProgressState(201)
	if err := NewJSONProgressStore(legacyFile).Save(want); err != nil {
		t.Fatalf("save legacy progress failed: %v", err)
	}

	stores, err := NewSQLiteStateStores(SQLiteStateStoresConfig{
		StateDBFile:        filepath.Join(dir, "sync-state.sqlite"),
		LegacyProgressFile: legacyFile,
	})
	if err != nil {
		t.Fatalf("create sqlite stores failed: %v", err)
	}

	got, err := stores.ProgressStore.Load()
	if err != nil {
		t.Fatalf("load progress failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected imported progress: %#v", got)
	}

	if err := os.Remove(legacyFile); err != nil {
		t.Fatalf("remove legacy progress failed: %v", err)
	}

	got, err = stores.ProgressStore.Load()
	if err != nil {
		t.Fatalf("reload progress from sqlite failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected persisted sqlite progress after legacy removal, got %#v", got)
	}
}

func TestSQLiteSyncStateStore_LegacyImport_PersistsIntoStateDB(t *testing.T) {
	dir := t.TempDir()
	legacyFile := filepath.Join(dir, "legacy-sync-state.json")
	want := sampleSyncState(1_720_000_000_000)
	if err := NewJSONSyncStateStore(legacyFile).Save(want); err != nil {
		t.Fatalf("save legacy sync state failed: %v", err)
	}

	stores, err := NewSQLiteStateStores(SQLiteStateStoresConfig{
		StateDBFile:         filepath.Join(dir, "sync-state.sqlite"),
		LegacySyncStateFile: legacyFile,
	})
	if err != nil {
		t.Fatalf("create sqlite stores failed: %v", err)
	}

	got, err := stores.SyncStateStore.Load()
	if err != nil {
		t.Fatalf("load sync state failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected imported sync state: %#v", got)
	}

	if err := os.Remove(legacyFile); err != nil {
		t.Fatalf("remove legacy sync state failed: %v", err)
	}

	got, err = stores.SyncStateStore.Load()
	if err != nil {
		t.Fatalf("reload sync state from sqlite failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected persisted sqlite sync state after legacy removal, got %#v", got)
	}
}

func TestSQLiteCheckpointStore_LegacyImport_PersistsIntoStateDB(t *testing.T) {
	dir := t.TempDir()
	checkpointKey := filepath.Join(dir, "legacy-checkpoint.json")
	want := sampleCheckpoint([]int64{5, 6, 7}, 15, 25)
	if err := NewJSONCheckpointStore(checkpointKey).Save(want); err != nil {
		t.Fatalf("save legacy checkpoint failed: %v", err)
	}

	stores, err := NewSQLiteStateStores(SQLiteStateStoresConfig{
		StateDBFile: filepath.Join(dir, "sync-state.sqlite"),
	})
	if err != nil {
		t.Fatalf("create sqlite stores failed: %v", err)
	}

	got, err := stores.CheckpointStoreFactory.ForKey(checkpointKey).Load()
	if err != nil {
		t.Fatalf("load checkpoint failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected imported checkpoint: %#v", got)
	}

	if err := os.Remove(checkpointKey); err != nil {
		t.Fatalf("remove legacy checkpoint failed: %v", err)
	}

	got, err = stores.CheckpointStoreFactory.ForKey(checkpointKey).Load()
	if err != nil {
		t.Fatalf("reload checkpoint from sqlite failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected persisted sqlite checkpoint after legacy removal, got %#v", got)
	}
}

func TestSQLiteProgressStore_ConcurrentSave_PreservesDecodableState(t *testing.T) {
	dir := t.TempDir()
	stores, err := NewSQLiteStateStores(SQLiteStateStoresConfig{
		StateDBFile: filepath.Join(dir, "sync-state.sqlite"),
	})
	if err != nil {
		t.Fatalf("create sqlite stores failed: %v", err)
	}

	const writers = 24
	var wg sync.WaitGroup
	errCh := make(chan error, writers)

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rootID := int64(i + 1)
			queueLen := int64(i + 3)
			state := &models.SyncProgressState{
				Status:    "running",
				Mode:      "full",
				StartedAt: int64(1_700_000_000_000 + i),
				UpdatedAt: int64(1_700_000_000_100 + i),
				Roots:     []int64{rootID},
				RootProgress: map[string]*models.RootSyncProgress{
					strconv.FormatInt(rootID, 10): &models.RootSyncProgress{
						RootFolderID:   rootID,
						CheckpointFile: fmt.Sprintf("checkpoint-%d.json", i),
						Status:         "running",
						QueueLength:    &queueLen,
						UpdatedAt:      int64(1_700_000_000_100 + i),
					},
				},
			}
			if err := stores.ProgressStore.Save(state); err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent save failed: %v", err)
		}
	}

	loaded, err := stores.ProgressStore.Load()
	if err != nil {
		t.Fatalf("load progress failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil progress after concurrent saves")
	}
	if loaded.Status == "" {
		t.Fatal("expected status to remain populated")
	}
	if len(loaded.Roots) == 0 {
		t.Fatal("expected roots to remain populated")
	}
	if len(loaded.RootProgress) == 0 {
		t.Fatal("expected root progress to remain populated")
	}

	for key, root := range loaded.RootProgress {
		if key == "" {
			t.Fatal("expected non-empty root progress key")
		}
		if root == nil {
			t.Fatal("expected non-nil root progress entry")
		}
		if root.RootFolderID == 0 {
			t.Fatal("expected root folder id to remain populated")
		}
		if root.CheckpointFile == "" {
			t.Fatal("expected checkpoint file to remain populated")
		}
		if root.QueueLength == nil {
			t.Fatal("expected queue length to remain populated")
		}
	}

	payload, err := json.Marshal(loaded)
	if err != nil {
		t.Fatalf("marshal loaded progress failed: %v", err)
	}

	var decoded models.SyncProgressState
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("re-decode loaded progress failed: %v", err)
	}
	if len(decoded.RootProgress) == 0 {
		t.Fatal("expected decoded root progress to remain populated")
	}
}

func sampleSyncProgressState(rootID int64) *models.SyncProgressState {
	queueLen := int64(4)
	currentFolderID := int64(rootID + 10)
	currentPageID := int64(rootID + 20)

	return &models.SyncProgressState{
		Status:             "running",
		Mode:               "full",
		StartedAt:          1_700_000_000_000,
		UpdatedAt:          1_700_000_000_123,
		MeiliHost:          "http://127.0.0.1:7700",
		MeiliIndex:         "npan_items",
		CheckpointTemplate: "./data/checkpoints/full-crawl.json",
		Roots:              []int64{rootID},
		CompletedRoots:     []int64{},
		RootProgress: map[string]*models.RootSyncProgress{
			strconv.FormatInt(rootID, 10): &models.RootSyncProgress{
				RootFolderID:    rootID,
				CheckpointFile:  filepath.Join("checkpoints", fmt.Sprintf("%d.json", rootID)),
				Status:          "running",
				CurrentFolderID: &currentFolderID,
				CurrentPageID:   &currentPageID,
				QueueLength:     &queueLen,
				UpdatedAt:       1_700_000_000_123,
			},
		},
	}
}

func sampleSyncState(lastSyncTime int64) *models.SyncState {
	return &models.SyncState{LastSyncTime: lastSyncTime}
}

func sampleCheckpoint(queue []int64, currentFolderID int64, currentPageID int64) *models.CrawlCheckpoint {
	return &models.CrawlCheckpoint{
		Queue:           queue,
		CurrentFolderID: &currentFolderID,
		CurrentPageID:   &currentPageID,
	}
}
