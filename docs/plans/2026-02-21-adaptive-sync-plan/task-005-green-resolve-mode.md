# Task 005: GREEN - resolveMode implementation

**depends-on**: task-002

## Description

Implement the `resolveMode` pure function that determines the sync mode based on the requested mode and current SyncState.

## Execution Context

**Task Number**: 005 of 010
**Phase**: Core Features
**Prerequisites**: Task 002 tests exist and fail

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 1 (auto → full), 2 (auto → incremental), 3 (explicit full), 4 (explicit incremental)

## Files to Modify/Create

- Modify: `internal/service/sync_manager.go`

## Steps

### Step 1: Implement resolveMode function

Add a `resolveMode` function to `sync_manager.go` that accepts `(mode models.SyncMode, state *models.SyncState)` and returns `models.SyncMode`.

Logic:
- If mode is "full" → return "full"
- If mode is "incremental" → return "incremental"
- If mode is "auto" or empty:
  - If state is non-nil AND state.LastSyncTime > 0 → return "incremental"
  - Otherwise → return "full"

### Step 2: Verify tests PASS (Green)

Run the mode resolution tests from Task 002 and verify they all pass.

### Step 3: Run full test suite

Ensure no regressions.

## Verification Commands

```bash
cd /root/workspace/npan && go test ./internal/service/ -run TestResolveMode -v
cd /root/workspace/npan && go test ./internal/service/ -v
```

## Success Criteria

- All TestResolveMode tests pass
- No test regressions
- Function is a pure function with no side effects
