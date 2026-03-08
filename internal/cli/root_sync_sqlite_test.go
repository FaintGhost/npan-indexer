package cli

import (
  "bytes"
  "io"
  "os"
  "path/filepath"
  "strings"
  "testing"

  "npan/internal/config"
  "npan/internal/models"
  "npan/internal/storage"
)

func TestSyncProgressCommand_ReadsProgressFromSQLiteWhenLegacyJSONMissing(t *testing.T) {
  t.Parallel()

  dir := t.TempDir()
  legacyProgressFile := filepath.Join(dir, "progress.json")
  stateDBFile := filepath.Join(dir, "sync-state.sqlite")

  stores, err := storage.NewSQLiteStateStores(storage.SQLiteStateStoresConfig{
    StateDBFile:        stateDBFile,
    LegacyProgressFile: legacyProgressFile,
  })
  if err != nil {
    t.Fatalf("create sqlite stores failed: %v", err)
  }
  t.Cleanup(func() {
    _ = stores.DB.Close()
  })

  if err := stores.ProgressStore.Save(&models.SyncProgressState{
    Status:         "done",
    Mode:           "full",
    StartedAt:      1700000000000,
    UpdatedAt:      1700000001000,
    Roots:          []int64{100},
    CompletedRoots: []int64{100},
    RootProgress:   map[string]*models.RootSyncProgress{},
  }); err != nil {
    t.Fatalf("save sqlite progress failed: %v", err)
  }

  cfg := config.Config{
    ProgressFile: legacyProgressFile,
    StateDBFile:  stateDBFile,
  }
  cmd := newSyncProgressCommand(cfg)
  cmd.SetArgs([]string{})

  output, err := captureStdout(func() error {
    return cmd.Execute()
  })
  if err != nil {
    t.Fatalf("sync-progress command failed: %v", err)
  }

  if strings.Contains(output, legacyProgressFile) && strings.Contains(output, "未找到") {
    t.Fatalf("expected command to read sqlite progress instead of failing on missing legacy file, got: %s", output)
  }
  if !strings.Contains(output, "\"status\": \"done\"") {
    t.Fatalf("expected sqlite-backed progress JSON output, got: %s", output)
  }
}

func captureStdout(run func() error) (string, error) {
  oldStdout := os.Stdout
  reader, writer, err := os.Pipe()
  if err != nil {
    return "", err
  }

  os.Stdout = writer
  defer func() {
    os.Stdout = oldStdout
  }()

  runErr := run()
  _ = writer.Close()

  var buf bytes.Buffer
  _, _ = io.Copy(&buf, reader)
  _ = reader.Close()

  return buf.String(), runErr
}
