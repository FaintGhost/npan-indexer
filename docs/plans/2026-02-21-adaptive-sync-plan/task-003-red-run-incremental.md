# Task 003: RED - runIncremental tests

**depends-on**: task-001

## Description

Write failing tests for the `runIncremental` method on SyncManager. These tests verify that incremental sync within SyncManager correctly handles retry on upsert/delete failures, tracks progress, and records incremental stats.

## Execution Context

**Task Number**: 003 of 010
**Phase**: Core Features
**Prerequisites**: Task 001 (model types must exist for compilation)

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 7 (retry on upsert), 8 (retry on delete), 9 (progress tracking)

## Files to Modify/Create

- Create: `internal/service/sync_manager_incremental_test.go`

## Steps

### Step 1: Create test file

Create `internal/service/sync_manager_incremental_test.go` in package `service`.

### Step 2: Define test helpers

Define mock/stub implementations for the dependencies that `runIncremental` will need:
- A mock `npan.API` that implements `SearchUpdatedWindow` returning pre-configured changes
- A mock `SyncStateStore` (using `indexer.SyncStateStore` interface)
- A mock `IndexWriter` for tracking upserted documents
- A mock `MeiliIndex` wrapper for delete operations
- Use the existing `models.RetryPolicyOptions` with `MaxRetries: 2` for tests

### Step 3: Test Scenario 7 — retry on upsert failure

Write a test where the upsert function fails with a retriable error on first call, then succeeds. Verify:
- The upsert is retried
- Progress state shows correct upserted count
- No skipped upserts

### Step 4: Test Scenario 7 — upsert exhausts retries

Write a test where upsert always fails with a retriable error. Verify:
- SkippedUpserts count matches the batch size
- Progress continues (does not terminate the sync)

### Step 5: Test Scenario 8 — retry on delete failure

Write a test where delete fails with a retriable error on first call, then succeeds. Verify:
- The delete is retried
- Progress state shows correct deleted count

### Step 6: Test Scenario 9 — progress tracking

Write a test with a successful incremental sync. Verify:
- SyncProgressState.Mode is "incremental"
- IncrementalStats is populated with ChangesFetched, Upserted, Deleted
- CursorBefore and CursorAfter are set correctly
- SyncState.LastSyncTime is updated after success

### Step 7: Verify tests FAIL (Red)

Run tests and verify they fail because `runIncremental` does not yet exist.

## Verification Commands

```bash
cd /root/workspace/npan && go test ./internal/service/ -run TestRunIncremental -v
# Expected: compilation error (runIncremental undefined)
```

## Success Criteria

- Test file created with 4+ test cases
- Tests cover retry, skip, progress tracking scenarios
- Tests use mock/stub dependencies (no real MeiliSearch or API calls)
- Tests fail because `runIncremental` does not exist
