# Task 001: Add model types and SyncManager dependency changes

**depends-on**: (none)

## Description

Add the foundational type definitions and structural changes needed by all subsequent tasks. This is purely additive — no behavior changes, just new types, fields, and dependency injection points.

## Execution Context

**Task Number**: 001 of 010
**Phase**: Foundation
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenario**: Supports all scenarios (type foundation)

## Files to Modify/Create

- Modify: `internal/models/models.go`
- Modify: `internal/service/sync_manager.go`

## Steps

### Step 1: Add SyncMode type to models

Add a `SyncMode` string type with three constants: `SyncModeAuto`, `SyncModeFull`, `SyncModeIncremental` to `internal/models/models.go`.

### Step 2: Add IncrementalSyncStats struct to models

Add `IncrementalSyncStats` struct to `internal/models/models.go` with fields:
- `ChangesFetched int64`
- `Upserted int64`
- `Deleted int64`
- `SkippedUpserts int64`
- `SkippedDeletes int64`
- `CursorBefore int64`
- `CursorAfter int64`

### Step 3: Add Mode and IncrementalStats fields to SyncProgressState

Add `Mode string` and `IncrementalStats *IncrementalSyncStats` (optional, omitempty) fields to the existing `SyncProgressState` struct.

### Step 4: Add Mode field to SyncStartRequest

Add `Mode models.SyncMode` field to `SyncStartRequest` in `internal/service/sync_manager.go`.

### Step 5: Add incremental fields to SyncStartRequest

Add `WindowOverlapMS int64` and `IncrementalQuery string` fields to `SyncStartRequest`.

### Step 6: Add SyncStateStore and incremental config to SyncManagerArgs

Add the following fields to `SyncManagerArgs` and `SyncManager`:
- `syncStateFile string` (path to sync state JSON file)
- `defaultIncrementalQuery string`
- `defaultWindowOverlapMS int64`

Update `NewSyncManager` to wire these fields.

### Step 7: Verify compilation

Ensure the project builds successfully with the new types.

## Verification Commands

```bash
cd /root/workspace/npan && go build ./...
```

## Success Criteria

- All new types and fields are defined
- Project compiles without errors
- No behavior changes — existing tests still pass
