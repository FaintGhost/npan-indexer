# Task 029: Implement frontend migration

**depends-on**: task-009

## Description

Migrate the frontend from `/demo` to `/app`. Rename the web directory, update HTML content to remove "demo" branding, and ensure the frontend works with the new `/api/v1/app/*` endpoints (no API key required).

## Execution Context

**Task Number**: 029 of 032
**Phase**: Frontend & Deployment
**Prerequisites**: Route restructure (task-009) must be complete so `/api/v1/app/*` routes exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 1 — public endpoint /app should be accessible without auth

## Files to Modify/Create

- Rename: `web/demo/index.html` → `web/app/index.html`
- Modify: `web/app/index.html` — update API endpoint URLs from `/api/v1/demo/*` to `/api/v1/app/*`, remove "demo" branding
- Modify: `internal/httpx/server.go` — update HTML path resolution (if not already done in task-009)

## Steps

### Step 1: Rename directory

- Move `web/demo/` to `web/app/`
- `git mv web/demo web/app`

### Step 2: Update frontend HTML

- In `web/app/index.html`:
  - Change API endpoint URLs from `/api/v1/demo/search` to `/api/v1/app/search`
  - Change API endpoint URLs from `/api/v1/demo/download-url` to `/api/v1/app/download-url`
  - Remove any "Demo" text from the title, headings, and branding
  - Remove any API key input fields (the embedded frontend doesn't need them)

### Step 3: Update path resolution

- In `internal/httpx/server.go`, ensure `resolveDemoHTMLPath` is renamed/updated to resolve `web/app/index.html` instead of `web/demo/index.html`

### Step 4: Update tests

- Update `server_demo_test.go` (or its replacement from task-009) to reference `/app` instead of `/demo`

### Step 5: Verify

- Run all httpx tests
- Manually verify `/app` serves the HTML page

## Verification Commands

```bash
go test ./internal/httpx/ -v
go test ./... -count=1
```

## Success Criteria

- `web/app/index.html` exists and is served at `/app`
- `web/demo/` no longer exists
- Frontend calls `/api/v1/app/*` endpoints
- No "demo" branding in the frontend
- All tests pass
