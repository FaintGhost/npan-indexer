# Task 002: Test config startup validation

**depends-on**: (none)

## Description

Write tests for config startup validation. The config `Validate()` method should reject empty AdminAPIKey, short AdminAPIKey (< 16 chars), missing required fields (MeiliHost, MeiliIndex, BaseURL), invalid numeric ranges, and incomplete auth credentials when fallback is enabled.

## Execution Context

**Task Number**: 002 of 032
**Phase**: Config Validation
**Prerequisites**: None — test task, can start immediately

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 1 "API Key 认证中间件" — Scenario "AdminAPIKey 为空时服务拒绝启动"

## Files to Modify/Create

- Create: `internal/config/validate_test.go`

## Steps

### Step 1: Verify Scenario

- Ensure Scenario "AdminAPIKey 为空时服务拒绝启动" exists in BDD specs

### Step 2: Implement Tests (Red)

- Create `internal/config/validate_test.go` with test cases:
  - `TestValidate_EmptyAdminAPIKey_ReturnsError` — empty AdminAPIKey should return error containing "NPA_ADMIN_API_KEY 不能为空"
  - `TestValidate_ShortAdminAPIKey_ReturnsError` — AdminAPIKey < 16 chars should return error about minimum length
  - `TestValidate_ValidConfig_NoError` — complete valid config returns nil
  - `TestValidate_MissingMeiliHost_ReturnsError` — empty MeiliHost returns error
  - `TestValidate_InvalidSyncConcurrency_ReturnsError` — SyncMaxConcurrent outside 1-20 returns error
  - `TestValidate_FallbackWithoutCredentials_ReturnsError` — AllowConfigAuthFallback=true without credentials returns error
- **Verification**: Run tests and verify they FAIL (Red) since `Validate()` doesn't exist yet

### Step 3: Verify Red State

- `go test ./internal/config/ -run TestValidate -v` should fail with compilation error (method not found)

## Verification Commands

```bash
go test ./internal/config/ -run TestValidate -v
```

## Success Criteria

- Test file compiles (once stub exists) and tests fail meaningfully
- Tests cover all validation rules from architecture.md Section 6
- Each test case maps to a specific validation rule
