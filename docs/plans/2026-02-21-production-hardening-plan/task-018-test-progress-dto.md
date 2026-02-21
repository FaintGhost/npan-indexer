# Task 018: Test progress response DTO

**depends-on**: (none)

## Description

Write tests verifying that the sync progress endpoint returns a sanitized DTO that excludes internal configuration fields (MeiliHost, MeiliIndex, CheckpointTemplate). The DTO should only contain operational fields relevant to the client.

## Execution Context

**Task Number**: 018 of 032
**Phase**: Error Handling
**Prerequisites**: None — test task

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 4 — "sync/full/progress 不泄露内部配置"

## Files to Modify/Create

- Create: `internal/httpx/dto_test.go`

## Steps

### Step 1: Verify Scenario

- Confirm scenario exists in Feature 4

### Step 2: Implement Tests (Red)

- Create `internal/httpx/dto_test.go` with:
  - `TestSyncProgressResponse_ExcludesMeiliHost` — convert a `SyncProgressState` (with MeiliHost set) to `SyncProgressResponse`; marshal to JSON; assert JSON does NOT contain "meiliHost"
  - `TestSyncProgressResponse_ExcludesMeiliIndex` — same for MeiliIndex
  - `TestSyncProgressResponse_ExcludesCheckpointTemplate` — same for CheckpointTemplate
  - `TestSyncProgressResponse_IncludesOperationalFields` — response includes status, startedAt, updatedAt, roots, aggregateStats
  - `TestRootProgressResponse_ExcludesCheckpointFile` — root progress excludes internal checkpoint file path
  - `TestRootProgressResponse_ExcludesRawError` — root progress excludes raw error string (uses sanitized lastError)
- Tests call a conversion function `toSyncProgressResponse(state *models.SyncProgressState) SyncProgressResponse`
- **Verification**: Tests FAIL (DTO types and conversion function don't exist)

## Verification Commands

```bash
go test ./internal/httpx/ -run TestSyncProgressResponse -v
go test ./internal/httpx/ -run TestRootProgressResponse -v
```

## Success Criteria

- Tests verify exclusion by checking marshaled JSON output
- Tests verify inclusion of needed operational fields
- Each internal field exclusion is explicitly tested
