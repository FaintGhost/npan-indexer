# Task 028: Implement io.LimitReader

**depends-on**: (none)

## Description

Add `io.LimitReader` to the npan API client to prevent unbounded memory consumption when reading upstream responses. Both error responses and normal responses should have size limits.

## Execution Context

**Task Number**: 028 of 032
**Phase**: Health & Ops
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: No direct BDD scenario
**Best practices ref**: `../2026-02-21-production-hardening-design/best-practices.md` #13

## Files to Modify/Create

- Modify: `internal/npan/client.go` — wrap `resp.Body` reads with `io.LimitReader`

## Steps

### Step 1: Limit error response reads

- In `client.go`, in the error response reading code (where `io.ReadAll(resp.Body)` is used for error bodies), wrap with `io.LimitReader(resp.Body, 4096)` — 4KB limit for error messages

### Step 2: Limit normal response reads

- For normal JSON response decoding, wrap with `io.LimitReader(resp.Body, 10*1024*1024)` — 10MB limit
- Use `json.NewDecoder(limited).Decode(out)` instead of reading all bytes first

### Step 3: Write test

- Add test to `internal/npan/client_test.go` (or create it):
  - `TestClient_LargeErrorResponse_Truncated` — mock HTTP server returns > 4KB error body; verify client handles it gracefully without OOM
  - `TestClient_LargeResponse_Truncated` — mock HTTP server returns > 10MB response; verify error is returned, not OOM

### Step 4: Verify

- Run tests: `go test ./internal/npan/ -v`

## Verification Commands

```bash
go test ./internal/npan/ -v
go vet ./internal/npan/
```

## Success Criteria

- All `io.ReadAll(resp.Body)` calls are wrapped with `io.LimitReader`
- Error responses limited to 4KB
- Normal responses limited to 10MB
- Tests verify truncation behavior
