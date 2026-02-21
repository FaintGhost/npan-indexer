# Task 006: GREEN - runIncremental implementation

**depends-on**: task-003

## Description

Implement the `runIncremental` method on SyncManager that handles incremental sync with retry, progress tracking, and stats within the SyncManager framework.

## Execution Context

**Task Number**: 006 of 010
**Phase**: Core Features
**Prerequisites**: Task 003 tests exist and fail

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 7 (retry upsert), 8 (retry delete), 9 (progress tracking)

## Files to Modify/Create

- Modify: `internal/service/sync_manager.go`

## Steps

### Step 1: Implement runIncremental method

Add a `runIncremental` method to SyncManager that:

1. Loads SyncState from the sync state file (create a `storage.JSONSyncStateStore` internally using `m.syncStateFile`)
2. Creates initial `SyncProgressState` with `Mode: "incremental"` and `IncrementalStats`
3. Calls `indexer.FetchIncrementalChanges` using the API's `SearchUpdatedWindow` method
4. Splits results into upserts and deletes
5. Applies upserts via `WithRetryVoid` through the `RequestLimiter`. On exhausted retries, increment `SkippedUpserts` and continue (do not terminate).
6. Applies deletes via `WithRetryVoid` through the `RequestLimiter`. On exhausted retries, increment `SkippedDeletes` and continue.
7. Updates `SyncProgressState.IncrementalStats` with counts
8. Persists progress to the progress store
9. On success, updates SyncState cursor to current time

The method signature should accept: `ctx context.Context, api npan.API, progress *models.SyncProgressState, request SyncStartRequest, limiter *indexer.RequestLimiter`.

### Step 2: Handle batch processing

For large change sets, batch upserts into groups (e.g., 200 per batch) and update progress after each batch. This matches how full crawl updates progress per page.

### Step 3: Handle context cancellation

Check `ctx.Err()` between batches. On cancellation, save current progress and return.

### Step 4: Verify tests PASS (Green)

Run the incremental tests from Task 003 and verify they all pass.

### Step 5: Run full test suite

Ensure no regressions.

## Verification Commands

```bash
cd /root/workspace/npan && go test ./internal/service/ -run TestRunIncremental -v
cd /root/workspace/npan && go test ./internal/service/ -v
```

## Success Criteria

- All TestRunIncremental tests pass
- Retry on upsert/delete failures works correctly
- Skipped items are tracked in IncrementalStats
- Progress is persisted during execution
- SyncState cursor is updated on success only
- No test regressions
