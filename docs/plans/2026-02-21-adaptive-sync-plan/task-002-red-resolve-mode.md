# Task 002: RED - Mode resolution tests

**depends-on**: task-001

## Description

Write failing tests for the `resolveMode` pure function that determines whether to run full or incremental sync based on mode parameter and SyncState cursor.

## Execution Context

**Task Number**: 002 of 010
**Phase**: Core Features
**Prerequisites**: Task 001 (model types must exist for compilation)

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 1 (first run → full), 2 (existing cursor → incremental), 3 (explicit full), 4 (explicit incremental with no cursor)

## Files to Modify/Create

- Create: `internal/service/sync_manager_mode_test.go`

## Steps

### Step 1: Create test file

Create `internal/service/sync_manager_mode_test.go` in package `service`.

### Step 2: Test Scenario 1 — auto mode, no sync state → full

Write a test that calls `resolveMode(models.SyncModeAuto, nil)` and asserts it returns `models.SyncModeFull`.

### Step 3: Test Scenario 2 — auto mode, existing cursor → incremental

Write a test that calls `resolveMode(models.SyncModeAuto, &models.SyncState{LastSyncTime: 1700000000})` and asserts it returns `models.SyncModeIncremental`.

### Step 4: Test Scenario 3 — explicit full overrides auto

Write a test that calls `resolveMode(models.SyncModeFull, &models.SyncState{LastSyncTime: 1700000000})` and asserts it returns `models.SyncModeFull`.

### Step 5: Test Scenario 4 — explicit incremental with no cursor

Write a test that calls `resolveMode(models.SyncModeIncremental, nil)` and asserts it returns `models.SyncModeIncremental`.

### Step 6: Test edge case — auto mode, zero cursor → full

Write a test that calls `resolveMode(models.SyncModeAuto, &models.SyncState{LastSyncTime: 0})` and asserts it returns `models.SyncModeFull`.

### Step 7: Verify tests FAIL (Red)

Run tests and verify they fail because `resolveMode` does not yet exist.

## Verification Commands

```bash
cd /root/workspace/npan && go test ./internal/service/ -run TestResolveMode -v
# Expected: compilation error (resolveMode undefined)
```

## Success Criteria

- Test file created with 5 test cases
- Tests fail because `resolveMode` function does not exist
- Test cases cover all 4 BDD scenarios plus edge case
