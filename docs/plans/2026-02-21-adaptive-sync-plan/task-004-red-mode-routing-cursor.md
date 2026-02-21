# Task 004: RED - Mode routing and cursor update tests

**depends-on**: task-001

## Description

Write failing tests for the mode routing logic in `SyncManager.run()` and the post-full-crawl cursor update behavior.

## Execution Context

**Task Number**: 004 of 010
**Phase**: Core Features
**Prerequisites**: Task 001 (model types must exist for compilation)

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 5 (full crawl updates cursor), 6 (failed full crawl does not update cursor), 10 (incremental verification), 11 (rate limiting)

## Files to Modify/Create

- Create: `internal/service/sync_manager_routing_test.go`

## Steps

### Step 1: Create test file

Create `internal/service/sync_manager_routing_test.go` in package `service`.

### Step 2: Test Scenario 5 — full crawl updates SyncState cursor

Write a test that runs a full crawl to completion using mocks. After completion, verify:
- `SyncState.LastSyncTime` has been written to the SyncStateStore
- The value is > 0 and represents a recent timestamp

### Step 3: Test Scenario 6 — failed full crawl does not update cursor

Write a test where the full crawl fails (mock API returns error). After failure, verify:
- SyncStateStore was NOT called with Save
- Or if it was, LastSyncTime is still 0/unchanged

### Step 4: Test Scenario 10 — incremental sync verification

Write a test that runs an incremental sync to completion. After completion, verify:
- `progress.Verification` is not nil
- `progress.Verification.MeiliDocCount` matches the mock value
- Verification is stored in the progress file

### Step 5: Test Scenario 11 — incremental uses rate limiter

Write a test that verifies the incremental path creates/uses a `RequestLimiter`. This can be verified by checking that upsert/delete operations are executed through the limiter's scheduling (verify ordering or timing constraints are respected).

### Step 6: Verify tests FAIL (Red)

Run tests and verify they fail because the routing logic and cursor update don't exist yet.

## Verification Commands

```bash
cd /root/workspace/npan && go test ./internal/service/ -run "TestRouting|TestCursorUpdate" -v
# Expected: compilation or assertion errors
```

## Success Criteria

- Test file created with 4 test cases
- Tests cover cursor write on success, no cursor on failure, verification, rate limiting
- Tests use mock dependencies
- Tests fail because implementation does not exist
