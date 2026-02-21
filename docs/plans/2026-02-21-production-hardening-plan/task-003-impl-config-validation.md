# Task 003: Implement config startup validation

**depends-on**: task-002

## Description

Implement the `Validate()` method on the `Config` struct that checks all required fields and security constraints. Integrate the validation call into `cmd/server/main.go` so the server refuses to start with invalid configuration.

## Execution Context

**Task Number**: 003 of 032
**Phase**: Config Validation
**Prerequisites**: Tests from task-002 must exist

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 1 — "AdminAPIKey 为空时服务拒绝启动"

## Files to Modify/Create

- Create: `internal/config/validate.go`
- Modify: `cmd/server/main.go` — add `cfg.Validate()` call after `config.Load()`

## Steps

### Step 1: Implement Validate method

- Create `internal/config/validate.go` with `func (c Config) Validate() error`
- Implement validation rules as described in architecture.md Section 6:
  - AdminAPIKey: not empty, minimum 16 characters
  - MeiliHost, MeiliIndex, BaseURL: not empty
  - SyncMaxConcurrent: 1-20 range
  - Retry.MaxRetries: 0-10 range
  - AllowConfigAuthFallback: if true, must have either client credentials or token
- Return aggregated error message with all failures listed

### Step 2: Integrate into server startup

- In `cmd/server/main.go`, after `config.Load()`, call `cfg.Validate()`
- If validation fails, log the error with `slog.Error` and exit with non-zero code
- Do NOT use `log.Fatal` — use `slog.Error` then `os.Exit(1)` for structured logging

### Step 3: Verify (Green)

- Run the tests from task-002 and verify they all PASS
- **Verification**: `go test ./internal/config/ -run TestValidate -v`

## Verification Commands

```bash
go test ./internal/config/ -run TestValidate -v
go test ./... -count=1
```

## Success Criteria

- All tests from task-002 pass
- Server refuses to start when AdminAPIKey is empty
- Validation errors are clear and actionable
- No regression in existing tests
