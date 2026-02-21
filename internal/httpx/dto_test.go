package httpx

import (
  "encoding/json"
  "strings"
  "testing"

  "npan/internal/models"
)

// --- TestToSyncProgressResponse_ExcludesMeiliHost ---

func TestToSyncProgressResponse_ExcludesMeiliHost(t *testing.T) {
  state := &models.SyncProgressState{
    Status:    "running",
    MeiliHost: "http://internal:7700",
  }

  resp := toSyncProgressResponse(state)

  data, err := json.Marshal(resp)
  if err != nil {
    t.Fatalf("json.Marshal failed: %v", err)
  }

  body := string(data)
  if strings.Contains(body, "meiliHost") {
    t.Errorf("response JSON must not contain field 'meiliHost', got: %s", body)
  }
  if strings.Contains(body, "http://internal:7700") {
    t.Errorf("response JSON must not contain the meili host value, got: %s", body)
  }
}

// --- TestToSyncProgressResponse_ExcludesMeiliIndex ---

func TestToSyncProgressResponse_ExcludesMeiliIndex(t *testing.T) {
  state := &models.SyncProgressState{
    Status:     "running",
    MeiliIndex: "my-secret-index",
  }

  resp := toSyncProgressResponse(state)

  data, err := json.Marshal(resp)
  if err != nil {
    t.Fatalf("json.Marshal failed: %v", err)
  }

  body := string(data)
  if strings.Contains(body, "meiliIndex") {
    t.Errorf("response JSON must not contain field 'meiliIndex', got: %s", body)
  }
  if strings.Contains(body, "my-secret-index") {
    t.Errorf("response JSON must not contain the meili index value, got: %s", body)
  }
}

// --- TestToSyncProgressResponse_ExcludesCheckpointTemplate ---

func TestToSyncProgressResponse_ExcludesCheckpointTemplate(t *testing.T) {
  state := &models.SyncProgressState{
    Status:             "running",
    CheckpointTemplate: "data/checkpoints/root-{id}.json",
  }

  resp := toSyncProgressResponse(state)

  data, err := json.Marshal(resp)
  if err != nil {
    t.Fatalf("json.Marshal failed: %v", err)
  }

  body := string(data)
  if strings.Contains(body, "checkpointTemplate") {
    t.Errorf("response JSON must not contain field 'checkpointTemplate', got: %s", body)
  }
  if strings.Contains(body, "data/checkpoints/root-{id}.json") {
    t.Errorf("response JSON must not contain the checkpoint template value, got: %s", body)
  }
}

// --- TestToSyncProgressResponse_IncludesOperationalFields ---

func TestToSyncProgressResponse_IncludesOperationalFields(t *testing.T) {
  roots := []int64{100, 200}
  completedRoots := []int64{100}
  activeRoot := int64(200)

  state := &models.SyncProgressState{
    Status:         "running",
    StartedAt:      1700000000,
    UpdatedAt:      1700001000,
    Roots:          roots,
    CompletedRoots: completedRoots,
    ActiveRoot:     &activeRoot,
    LastError:      "some error",
  }

  resp := toSyncProgressResponse(state)

  data, err := json.Marshal(resp)
  if err != nil {
    t.Fatalf("json.Marshal failed: %v", err)
  }

  body := string(data)
  if !strings.Contains(body, `"running"`) {
    t.Errorf("expected status 'running' in response, got: %s", body)
  }
  if !strings.Contains(body, "1700000000") {
    t.Errorf("expected startedAt in response, got: %s", body)
  }
  if !strings.Contains(body, "100") || !strings.Contains(body, "200") {
    t.Errorf("expected roots in response, got: %s", body)
  }
}

// --- TestToRootProgressResponse_ExcludesCheckpointFile ---

func TestToRootProgressResponse_ExcludesCheckpointFile(t *testing.T) {
  rp := &models.RootSyncProgress{
    RootFolderID:   42,
    Status:         "done",
    CheckpointFile: "/data/checkpoints/root-42.json",
  }

  resp := toRootProgressResponse(rp)

  data, err := json.Marshal(resp)
  if err != nil {
    t.Fatalf("json.Marshal failed: %v", err)
  }

  body := string(data)
  if strings.Contains(body, "checkpointFile") {
    t.Errorf("response JSON must not contain field 'checkpointFile', got: %s", body)
  }
  if strings.Contains(body, "/data/checkpoints/root-42.json") {
    t.Errorf("response JSON must not contain the checkpoint file path, got: %s", body)
  }
}

// --- TestToRootProgressResponse_ExcludesRawError ---

func TestToRootProgressResponse_ExcludesRawError(t *testing.T) {
  rp := &models.RootSyncProgress{
    RootFolderID: 42,
    Status:       "error",
    Error:        "internal connection timeout to 10.0.0.1:7700",
  }

  resp := toRootProgressResponse(rp)

  data, err := json.Marshal(resp)
  if err != nil {
    t.Fatalf("json.Marshal failed: %v", err)
  }

  body := string(data)
  if strings.Contains(body, "internal connection timeout to 10.0.0.1:7700") {
    t.Errorf("response JSON must not contain the raw error string, got: %s", body)
  }
}
