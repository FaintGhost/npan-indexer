# Task 012: Test type parameter whitelist

**depends-on**: (none)

## Description

Write tests for type parameter whitelist validation. The type parameter in search queries should only accept "all", "file", "folder". Any other value, including injection attempts, should be rejected with 400.

## Execution Context

**Task Number**: 012 of 032
**Phase**: Input Validation
**Prerequisites**: None — test task

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 3 — Scenario Outline "type 参数只接受白名单值"

## Files to Modify/Create

- Modify: `internal/httpx/validation_test.go` (append to file from task-010)

## Steps

### Step 1: Verify Scenario

- Confirm Scenario Outline with injection examples exists

### Step 2: Implement Tests (Red)

- Add to `internal/httpx/validation_test.go`:
  - `TestValidateType_AllowedValues` — table-driven test: "all" → ok, "file" → ok, "folder" → ok
  - `TestValidateType_InjectionAttempt_Returns400` — "file OR is_deleted = true" → error
  - `TestValidateType_SQLInjection_Returns400` — "' OR 1=1 --" → error
  - `TestValidateType_EmptyString_Allowed` — empty string → ok (means "all")
  - `TestValidateType_ArbitraryString_Returns400` — "invalid" → error
- Tests call `validateType(value string) error`
- **Verification**: Tests FAIL (function doesn't exist)

## Verification Commands

```bash
go test ./internal/httpx/ -run TestValidateType -v
```

## Success Criteria

- All BDD scenario outline examples covered
- Injection attempts are explicitly tested
- Empty type treated as "all" (default)
