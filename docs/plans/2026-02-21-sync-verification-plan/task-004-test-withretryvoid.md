# Task 004: Test WithRetryVoid

**depends-on**: none

## BDD Reference

- Scenario: Transient MeiliSearch failure is retried (uses WithRetryVoid)

## Description

Add tests to `internal/indexer/retry_test.go`:

1. `TestWithRetryVoid_Success` — operation succeeds on first try, verify no error returned
2. `TestWithRetryVoid_RetryThenSuccess` — operation fails with retriable error, then succeeds. Verify 2 attempts and no error.
3. `TestWithRetryVoid_PermanentFailure` — operation fails with non-retriable error. Verify 1 attempt and error returned.

These tests should initially FAIL since `WithRetryVoid` does not exist yet.

## Files

- `internal/indexer/retry_test.go` — modify (add test cases)

## Verification

```bash
go test ./internal/indexer/ -run "TestWithRetryVoid" -v
```

Expect: compilation error / FAIL (Red phase).
