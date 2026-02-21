# Task 007: GREEN - Mode routing and cursor update

**depends-on**: task-004, task-005, task-006

## Description

Modify `SyncManager.run()` to use `resolveMode` for mode selection and dispatch to either the existing full crawl path or the new `runIncremental` path. Add cursor update after successful full crawl.

## Execution Context

**Task Number**: 007 of 010
**Phase**: Integration
**Prerequisites**: Tasks 004 (tests), 005 (resolveMode), 006 (runIncremental) all complete

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 5 (cursor update), 6 (no cursor on failure), 10 (incremental verification), 11 (rate limiting)

## Files to Modify/Create

- Modify: `internal/service/sync_manager.go`

## Steps

### Step 1: Add mode resolution to run()

At the top of `SyncManager.run()`, after discovering root folders:
1. Load SyncState from the sync state file
2. Call `resolveMode(request.Mode, syncState)` to get the effective mode
3. Set `progress.Mode` to the resolved mode string

### Step 2: Add mode dispatch

After mode resolution:
- If resolved mode is "full": continue with existing full crawl logic (unchanged)
- If resolved mode is "incremental": call `runIncremental()` instead

### Step 3: Add cursor update after successful full crawl

After the existing full crawl completes successfully (after `progress.Status = "done"`):
1. Create a `storage.JSONSyncStateStore` for the sync state file
2. Write `SyncState{LastSyncTime: time.Now().Unix()}` to enable future incremental runs

### Step 4: Add verification for incremental mode

The `runIncremental` method should call `buildVerification()` after completion (similar to how full crawl does it in `run()`). If verification is done inside `runIncremental`, ensure the progress is saved with verification data.

### Step 5: Verify routing tests PASS (Green)

Run the routing and cursor tests from Task 004.

### Step 6: Run full test suite

Ensure no regressions. Verify all existing full sync tests still pass.

## Verification Commands

```bash
cd /root/workspace/npan && go test ./internal/service/ -run "TestRouting|TestCursorUpdate" -v
cd /root/workspace/npan && go test ./internal/service/ -v
cd /root/workspace/npan && go test ./... 2>&1 | tail -20
```

## Success Criteria

- Mode routing dispatches correctly to full or incremental
- Full crawl writes SyncState cursor on success
- Failed full crawl does not write cursor
- Incremental sync includes verification
- All existing tests continue to pass
- No test regressions
