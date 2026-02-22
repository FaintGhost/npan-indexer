package httpx

import "npan/internal/models"

type SyncProgressResponse struct {
  Status           string                           `json:"status"`
  Mode             string                           `json:"mode,omitempty"`
  StartedAt        int64                            `json:"startedAt"`
  UpdatedAt        int64                            `json:"updatedAt"`
  Roots            []int64                          `json:"roots"`
  RootNames        map[int64]string                 `json:"rootNames,omitempty"`
  CompletedRoots   []int64                          `json:"completedRoots"`
  ActiveRoot       *int64                           `json:"activeRoot,omitempty"`
  AggregateStats   models.CrawlStats                `json:"aggregateStats"`
  RootProgress     map[string]*RootProgressResponse `json:"rootProgress"`
  IncrementalStats *models.IncrementalSyncStats      `json:"incrementalStats,omitempty"`
  LastError        string                           `json:"lastError,omitempty"`
  Verification     *models.SyncVerification          `json:"verification,omitempty"`
}

type RootProgressResponse struct {
  RootFolderID       int64             `json:"rootFolderId"`
  Status             string            `json:"status"`
  EstimatedTotalDocs *int64            `json:"estimatedTotalDocs,omitempty"`
  Stats              models.CrawlStats `json:"stats"`
  UpdatedAt          int64             `json:"updatedAt"`
}

func toSyncProgressResponse(state *models.SyncProgressState) SyncProgressResponse {
  resp := SyncProgressResponse{
    Status:           state.Status,
    Mode:             state.Mode,
    StartedAt:        state.StartedAt,
    UpdatedAt:        state.UpdatedAt,
    Roots:            state.Roots,
    RootNames:        state.RootNames,
    CompletedRoots:   state.CompletedRoots,
    ActiveRoot:       state.ActiveRoot,
    AggregateStats:   state.AggregateStats,
    IncrementalStats: state.IncrementalStats,
    LastError:        state.LastError,
    Verification:     state.Verification,
  }
  if state.RootProgress != nil {
    resp.RootProgress = make(map[string]*RootProgressResponse, len(state.RootProgress))
    for k, v := range state.RootProgress {
      rpr := toRootProgressResponse(v)
      resp.RootProgress[k] = &rpr
    }
  }
  return resp
}

func toRootProgressResponse(rp *models.RootSyncProgress) RootProgressResponse {
  return RootProgressResponse{
    RootFolderID:       rp.RootFolderID,
    Status:             rp.Status,
    EstimatedTotalDocs: rp.EstimatedTotalDocs,
    Stats:              rp.Stats,
    UpdatedAt:          rp.UpdatedAt,
  }
}
