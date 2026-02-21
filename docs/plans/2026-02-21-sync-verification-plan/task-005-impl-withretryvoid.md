# Task 005: Implement WithRetryVoid

**depends-on**: task-004

## BDD Reference

- Scenario: Transient MeiliSearch failure is retried

## Description

Add `WithRetryVoid` function to `internal/indexer/retry.go`:

- Signature: `func WithRetryVoid(ctx context.Context, operation func() error, opts models.RetryPolicyOptions) error`
- Implementation: wrap `operation` into `func() (struct{}, error)` and delegate to existing `WithRetry`
- Return only the error

## Files

- `internal/indexer/retry.go` — modify (add function)

## Verification

```bash
go test ./internal/indexer/ -run "TestWithRetryVoid" -v
```

Expect: all 3 tests PASS (Green phase).
