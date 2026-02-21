# Adaptive Sync Design

## Context

The current codebase has two completely separate sync paths:

1. **sync-full** (`SyncManager` + `RunFullCrawl`): BFS directory traversal via `ListFolderChildren`. Has rich progress tracking, per-root checkpoints, HTTP API, retry, rate limiting, and 4-layer verification.

2. **sync-incremental** (`RunIncrementalSync` + `FetchIncrementalChanges`): Time-window query via `SearchUpdatedWindow`. CLI-only, no retry on write, no verification, no rate limiting, no HTTP endpoint.

Users must manually decide which command to run and remember the correct flags for each.

## Requirements

1. **Single entry point**: One `sync` command (CLI) and one `/api/v1/admin/sync` endpoint (HTTP) that automatically selects the right mode.
2. **Auto-detection**: First run or missing state -> full crawl. State exists with valid cursor -> incremental. Explicit `mode` override supported.
3. **Feature parity**: Incremental sync gains the same production-quality features as full sync: retry on writes, rate limiting, progress tracking via `SyncProgressState`, verification, and HTTP API.
4. **Delete handling**: Incremental handles deletes (already does). Full sync does not need to change (it's additive by nature).
5. **Backward compatibility**: Old `sync-full` and `sync-incremental` CLI commands remain as aliases.
6. **Post-sync cursor update**: After a successful full crawl, write `SyncState.LastSyncTime` so the next auto-detected run will be incremental.

## Rationale

### Why absorb into SyncManager (not a thin router)?

- SyncManager already owns progress tracking, cancellation, mutex-based run guard, and HTTP wiring.
- A thin router would duplicate the "is running?" check, progress store access, and cancellation logic.
- Keeping one orchestrator means one place to add features (observability, webhooks, etc.).

### Why not merge the crawl engines?

- `ListFolderChildren` (BFS) and `SearchUpdatedWindow` (time-window) are fundamentally different APIs.
- They return different data shapes and pagination models.
- Merging them would create unnecessary complexity. Better to keep them as separate strategies behind a unified orchestrator.

## Detailed Design

### 1. Sync Mode Enum

```go
type SyncMode string

const (
  SyncModeAuto        SyncMode = "auto"
  SyncModeFull        SyncMode = "full"
  SyncModeIncremental SyncMode = "incremental"
)
```

### 2. SyncStartRequest Changes

```go
type SyncStartRequest struct {
  Mode               SyncMode `json:"mode"` // NEW: "auto" | "full" | "incremental"
  // ... existing fields for full crawl ...
  // NEW incremental fields:
  WindowOverlapMS    int64    `json:"window_overlap_ms"`
  IncrementalQuery   string   `json:"incremental_query"`
}
```

### 3. SyncManager.run() Decision Logic

```
if mode == "full":
    run full crawl (existing path)
elif mode == "incremental":
    run incremental (new path in SyncManager)
elif mode == "auto":
    load SyncState
    if SyncState exists and LastSyncTime > 0:
        run incremental
    else:
        run full crawl
```

### 4. Incremental Path in SyncManager

A new `runIncremental()` method that:
- Uses the shared `RequestLimiter` for rate limiting
- Wraps upsert/delete with `WithRetryVoid` for retry
- Tracks progress in `SyncProgressState` (with `mode: "incremental"` field)
- Calls `buildVerification()` post-sync
- Updates `SyncState.LastSyncTime` on success

### 5. SyncProgressState Changes

```go
type SyncProgressState struct {
  Mode string `json:"mode"` // NEW: "full" | "incremental"
  // ... existing fields ...
  // NEW incremental-specific stats:
  IncrementalStats *IncrementalSyncStats `json:"incrementalStats,omitempty"`
}

type IncrementalSyncStats struct {
  ChangesFetched int64 `json:"changesFetched"`
  Upserted       int64 `json:"upserted"`
  Deleted        int64 `json:"deleted"`
  SkippedUpserts int64 `json:"skippedUpserts"`
  SkippedDeletes int64 `json:"skippedDeletes"`
  CursorBefore   int64 `json:"cursorBefore"`
  CursorAfter    int64 `json:"cursorAfter"`
}
```

### 6. Post-Full-Crawl Cursor Write

After successful full crawl, `SyncManager.run()` writes `SyncState{LastSyncTime: now}` to the sync state file. This enables auto-detection to choose incremental on the next run.

### 7. SyncManager Dependencies Update

```go
type SyncManagerArgs struct {
  // ... existing fields ...
  SyncStateStore     storage.JSONSyncStateStore // NEW: for incremental cursor
  IncrementalQuery   string                     // NEW: default query words
  WindowOverlapMS    int64                      // NEW: default overlap
}
```

### 8. CLI Changes

New `sync` command (primary):
```
npan-cli sync [--mode auto|full|incremental] [existing flags...]
```

Old commands become aliases:
- `sync-full` -> `sync --mode full`
- `sync-incremental` -> `sync --mode incremental`

### 9. HTTP Changes

New endpoint: `POST /api/v1/admin/sync/start` (or reuse `/sync/full` with mode field)
- Request body includes `mode` field
- Response unchanged

Old endpoints remain as aliases for backward compatibility.

### 10. Frontend Changes

- `SyncProgressDisplay` shows mode indicator ("full" / "incremental")
- Incremental mode shows different stat cards (changes/upserts/deletes instead of folders/pages)
- Verification section works for both modes

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
