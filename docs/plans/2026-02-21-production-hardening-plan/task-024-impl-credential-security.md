# Task 024: Implement credential security

**depends-on**: task-023

## Description

Implement config log sanitization via `slog.LogValuer` interface and remove the query parameter token fallback from handlers. This addresses two security issues: credential leakage in logs and token exposure in URLs.

## Execution Context

**Task Number**: 024 of 032
**Phase**: Credential Security
**Prerequisites**: Config log tests (task-023) must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 5 — "日志不打印敏感字段"
**Additional ref**: best-practices.md #4 (remove query parameter token)

## Files to Modify/Create

- Modify: `internal/config/config.go` — add `LogValue()` method
- Modify: `internal/httpx/handlers.go` — remove `c.QueryParam("token")` from `resolveAuthOptions`

## Steps

### Step 1: Implement LogValue

- Add `func (c Config) LogValue() slog.Value` to `internal/config/config.go`
- Return `slog.GroupValue` with all fields; sensitive fields (AdminAPIKey, ClientSecret, MeiliAPIKey, Token) replaced with `[REDACTED]`
- Non-sensitive fields (ServerAddr, BaseURL, MeiliHost, MeiliIndex) show actual values

### Step 2: Remove query parameter token

- In `internal/httpx/handlers.go`, in `resolveAuthOptions` (around line 104), remove the `c.QueryParam("token")` line
- Token should only be accepted via request body or header, never via URL query parameter

### Step 3: Verify (Green)

- Run log sanitization tests from task-023
- Run existing handler tests to ensure no regression
- **Verification**: `go test ./internal/config/ -run TestConfig_LogValue -v`

## Verification Commands

```bash
go test ./internal/config/ -run TestConfig_LogValue -v
go test ./internal/httpx/ -v
go test ./... -count=1
```

## Success Criteria

- Config log sanitization tests pass
- Query parameter token removed from resolveAuthOptions
- No regression in handler tests
