# Task 009: HTTP API update

**depends-on**: task-001

## Description

Update the HTTP handler and server routes to support the unified sync API with mode parameter.

## Execution Context

**Task Number**: 009 of 010
**Phase**: Integration
**Prerequisites**: Task 001 (SyncMode type and SyncStartRequest.Mode field exist)

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 15 (POST /sync/start accepts mode), 16 (GET /progress returns mode)

## Files to Modify/Create

- Modify: `internal/httpx/handlers.go`
- Modify: `internal/httpx/server.go`
- Modify: `internal/httpx/handlers_test.go`

## Steps

### Step 1: Update syncStartPayload

Add `Mode string` field to `syncStartPayload` struct in handlers.go. Also add incremental-specific fields: `WindowOverlapMS`, `IncrementalQuery`.

### Step 2: Update StartFullSync handler

Modify `StartFullSync` to pass the mode from payload to `SyncStartRequest`. If mode is empty, default to "auto" (backward compatible — existing clients that don't send mode get auto behavior, which without a cursor falls back to full).

### Step 3: Add unified route

In `server.go`, add a new route `admin.POST("/sync/start", handlers.StartFullSync)` that points to the same handler. Keep the existing `/sync/full` route as an alias.

### Step 4: Progress endpoint already returns mode

The `GetFullSyncProgress` handler returns the full `SyncProgressState` JSON, which now includes the `mode` field (added in Task 001). No handler change needed, but verify the response includes it.

### Step 5: Write test for mode in request

Add a test in `handlers_test.go` that sends a POST to `/api/v1/admin/sync/start` with `{"mode": "auto"}` and verifies the sync starts successfully. Add another test with `{"mode": "incremental"}`.

### Step 6: Write test for mode in progress response

Add a test that verifies the progress response includes the `mode` field.

### Step 7: Verify

Run handler tests and ensure all pass.

## Verification Commands

```bash
cd /root/workspace/npan && go test ./internal/httpx/ -run TestSync -v
cd /root/workspace/npan && go test ./internal/httpx/ -v
```

## Success Criteria

- `POST /api/v1/admin/sync/start` accepts mode field
- `POST /api/v1/admin/sync/full` still works (backward compatible)
- Progress response includes mode field
- Handler tests pass
- No regressions in existing handler tests
