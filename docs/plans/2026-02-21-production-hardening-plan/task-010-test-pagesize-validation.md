# Task 010: Test pageSize validation

**depends-on**: (none)

## Description

Write tests for pageSize input validation. The handler should reject pageSize values exceeding 100 and return 400. The query service should also cap pageSize at 100 as a defense-in-depth measure.

## Execution Context

**Task Number**: 010 of 032
**Phase**: Input Validation
**Prerequisites**: None — test task

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**:
- "pageSize 超过上限返回 400"
- "pageSize 为 0 或负数返回 400"
- "pageSize 在有效范围内正常返回"

## Files to Modify/Create

- Create: `internal/httpx/validation_test.go`

## Steps

### Step 1: Verify Scenarios

- Confirm all 3 pageSize scenarios exist in Feature 3

### Step 2: Implement Tests (Red)

- Create `internal/httpx/validation_test.go` with:
  - `TestValidatePageSize_ExceedsMax_Returns400` — pageSize=1001 returns error
  - `TestValidatePageSize_Negative_Returns400` — pageSize=-1 returns error
  - `TestValidatePageSize_Zero_Returns400` — pageSize=0 returns error
  - `TestValidatePageSize_ValidRange_NoError` — pageSize=50 returns no error
  - `TestValidatePageSize_MaxBoundary_NoError` — pageSize=100 returns no error
  - `TestValidatePageSize_OverMaxBoundary_ReturnsError` — pageSize=101 returns error
- Tests should call a `validatePageSize(value int64) error` function
- **Verification**: Tests FAIL since validation function doesn't exist

## Verification Commands

```bash
go test ./internal/httpx/ -run TestValidatePageSize -v
```

## Success Criteria

- Tests cover all 3 BDD scenarios plus boundary cases
- Tests verify error messages are user-friendly
