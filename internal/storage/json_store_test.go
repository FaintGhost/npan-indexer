package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"npan/internal/models"
)

func TestWriteFileAtomic_ReplacesContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	if err := writeFileAtomic(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("write 1 failed: %v", err)
	}
	if err := writeFileAtomic(path, []byte("second"), 0o644); err != nil {
		t.Fatalf("write 2 failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "second" {
		t.Fatalf("expected second, got %q", string(data))
	}
}

func TestJSONProgressStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "progress.json")
	store := NewJSONProgressStore(path)

	input := &models.SyncProgressState{
		Status:         "running",
		Roots:          []int64{0, 1},
		CompletedRoots: []int64{0},
	}

	if err := store.Save(input); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil progress")
	}
	if loaded.Status != "running" {
		t.Fatalf("unexpected status: %q", loaded.Status)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp-") {
			t.Fatalf("unexpected leftover temp file: %s", entry.Name())
		}
	}
}
