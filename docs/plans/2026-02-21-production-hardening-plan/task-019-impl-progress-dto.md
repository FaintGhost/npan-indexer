# Task 019: Implement progress response DTO

**depends-on**: task-018

## Description

Create `SyncProgressResponse` and `RootProgressResponse` DTO types and a conversion function. Integrate into the `GetFullSyncProgress` handler to return the sanitized DTO instead of the raw internal state.

## Execution Context

**Task Number**: 019 of 032
**Phase**: Error Handling
**Prerequisites**: Progress DTO tests (task-018) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 4 — "sync/full/progress 不泄露内部配置"

## Files to Modify/Create

- Create: `internal/httpx/dto.go`
- Modify: `internal/httpx/handlers.go` — update `GetFullSyncProgress` to use DTO conversion

## Steps

### Step 1: Create DTO types

- Create `internal/httpx/dto.go` with:
  - `SyncProgressResponse` struct — fields: Status, StartedAt, UpdatedAt, Roots, CompletedRoots, ActiveRoot, AggregateStats, RootProgress, LastError (sanitized)
  - `RootProgressResponse` struct — fields: RootFolderID, Status, EstimatedTotalDocs, Stats, UpdatedAt
  - `func toSyncProgressResponse(state *models.SyncProgressState) SyncProgressResponse` — conversion function that maps fields and excludes internal config

### Step 2: Integrate into handler

- In `GetFullSyncProgress` handler, after getting the progress state, call `toSyncProgressResponse` and return the DTO instead of the raw state

### Step 3: Verify (Green)

- Run tests from task-018
- **Verification**: `go test ./internal/httpx/ -run TestSyncProgressResponse -v`

## Verification Commands

```bash
go test ./internal/httpx/ -run TestSyncProgressResponse -v
go test ./internal/httpx/ -run TestRootProgressResponse -v
go test ./... -count=1
```

## Success Criteria

- All DTO tests pass
- MeiliHost, MeiliIndex, CheckpointTemplate not present in API response
- Operational fields correctly mapped
