# Architecture: Adaptive Sync

## Component Diagram

```
                         ┌─────────────────────┐
                         │   CLI: sync command  │
                         │  (auto/full/incr)    │
                         └──────────┬───────────┘
                                    │
                         ┌──────────▼───────────┐
                         │     HTTP Handler      │
                         │ POST /admin/sync/start│
                         │ GET /admin/sync/progress │
                         │ POST /admin/sync/cancel  │
                         └──────────┬───────────┘
                                    │
                         ┌──────────▼───────────┐
                         │    SyncManager        │
                         │  (unified orchestrator)│
                         │                       │
                         │  ┌─────────────────┐  │
                         │  │  Mode Resolver   │  │
                         │  │  auto→full/incr  │  │
                         │  └────────┬────────┘  │
                         │           │           │
                         │  ┌────────▼────────┐  │
                         │  │  Progress Store  │  │
                         │  │ (SyncProgressState)│  │
                         │  └────────┬────────┘  │
                         │           │           │
                         │    ┌──────┴──────┐    │
                         │    │             │    │
                    ┌────▼────▼──┐  ┌───────▼───────┐
                    │ Full Crawl  │  │ Incremental   │
                    │ Strategy    │  │ Strategy      │
                    │             │  │               │
                    │ BFS via     │  │ Time-window   │
                    │ ListFolder  │  │ via Search    │
                    │ Children    │  │ UpdatedWindow │
                    └──────┬──────┘  └───────┬───────┘
                           │                 │
                    ┌──────▼─────────────────▼──────┐
                    │      Shared Infrastructure     │
                    │  RequestLimiter, WithRetry,    │
                    │  MeiliIndex, Verification      │
                    └────────────────────────────────┘
```

## Key Architectural Decisions

### 1. SyncManager as Unified Orchestrator

SyncManager already handles: mutex-based run guard, goroutine lifecycle, cancellation via context, progress persistence, and HTTP wiring. Rather than creating a parallel orchestrator for incremental, we extend SyncManager with a mode selector.

**Impact**: `SyncManager.run()` gains a mode switch at the top. The full crawl path is unchanged. A new `runIncremental()` method handles incremental logic.

### 2. Incremental Strategy Within SyncManager

The new `runIncremental()` method:

```
1. Load SyncState → get lastSyncTime cursor
2. Create initial SyncProgressState with mode="incremental"
3. Call FetchIncrementalChanges (existing function, unchanged)
4. Split into upserts/deletes (existing logic)
5. Apply upserts via shared limiter + WithRetryVoid
6. Apply deletes via shared limiter + WithRetryVoid
7. Update progress store after each batch
8. Run buildVerification
9. Update SyncState cursor
```

### 3. SyncState Store as SyncManager Dependency

Currently, `SyncStateStore` is created locally in the CLI command. To enable auto-detection and cursor updates in SyncManager, the `SyncStateStore` must be injected via `SyncManagerArgs`.

### 4. Incremental-Specific Fields in SyncProgressState

Instead of creating a separate progress model, add an `IncrementalStats` optional field. This keeps the frontend schema compatible — existing clients ignore unknown optional fields.

### 5. FetchChanges Closure Extraction

Currently, the incremental CLI command creates a complex closure for `FetchChanges` that wraps `FetchIncrementalChanges`. This closure will move into `SyncManager.runIncremental()` to keep all sync logic in the orchestrator.

The `api npan.API` parameter already available to `run()` provides the `SearchUpdatedWindow` method needed.

## File Changes Summary

### Modified Files

| File | Change |
|------|--------|
| `internal/models/models.go` | Add `SyncMode`, `IncrementalSyncStats`, mode field to `SyncProgressState` |
| `internal/service/sync_manager.go` | Add `runIncremental()`, mode resolver, SyncState cursor write post-full |
| `internal/service/sync_manager.go` | Update `SyncManagerArgs` with SyncStateStore and incremental config |
| `internal/service/sync_manager.go` | Update `SyncStartRequest` with Mode field |
| `internal/httpx/handlers.go` | Update `syncStartPayload` and handler to pass mode |
| `internal/httpx/server.go` | Add unified `/sync/start` route (keep old routes as aliases) |
| `internal/cli/root.go` | Add `sync` command, make old commands thin wrappers |
| `internal/config/config.go` | Already has `SyncStateFile`, `IncrementalQuery` — no change needed |
| `web/src/lib/sync-schemas.ts` | Add mode, incrementalStats fields |
| `web/src/components/sync-progress-display.tsx` | Conditional rendering based on mode |

### New Files

| File | Purpose |
|------|---------|
| `internal/service/sync_manager_incremental_test.go` | Tests for incremental path in SyncManager |
| `internal/service/sync_manager_mode_test.go` | Tests for auto-detection logic |

### Unchanged Files

| File | Reason |
|------|--------|
| `internal/indexer/incremental_sync.go` | Still used as the core incremental engine |
| `internal/indexer/incremental_fetch.go` | Still used for fetching changes |
| `internal/indexer/full_crawl.go` | Full crawl logic unchanged |
| `internal/indexer/retry.go` | Retry infrastructure unchanged |
| `internal/indexer/limiter.go` | Rate limiter unchanged |
| `internal/search/meili_index.go` | MeiliSearch operations unchanged |

## Data Flow: Auto Mode

```
sync(mode="auto")
  → SyncManager.run()
    → Load SyncState from JSONSyncStateStore
    → if state != nil && state.LastSyncTime > 0:
        resolvedMode = "incremental"
      else:
        resolvedMode = "full"
    → Set progress.Mode = resolvedMode
    → if resolvedMode == "full":
        existing full crawl path
        on success: write SyncState{LastSyncTime: now}
      elif resolvedMode == "incremental":
        runIncremental()
    → buildVerification()
    → Save progress
```
