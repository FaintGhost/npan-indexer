# Task 002: Test isRetriable for MeiliSearch errors

**depends-on**: none

## BDD Reference

- Scenario: MeiliSearch timeout is retriable
- Scenario: MeiliSearch 429 is retriable

## Description

Add tests to `internal/indexer/retry_test.go` that verify `isRetriable` correctly classifies MeiliSearch errors:

1. `TestIsRetriable_MeiliSearchTimeout` — create `*meilisearch.Error` with `ErrCode: MeilisearchTimeoutError`, assert `isRetriable` returns `true`
2. `TestIsRetriable_MeiliSearch429` — create `*meilisearch.Error` with `ErrCode: MeilisearchApiError` and `StatusCode: 429`, assert `true`
3. `TestIsRetriable_MeiliSearch503` — `StatusCode: 503`, assert `true`
4. `TestIsRetriable_MeiliSearch400` — `StatusCode: 400`, assert `false` (not retriable)
5. `TestIsRetriable_MeiliSearchCommunicationError` — `ErrCode: MeilisearchCommunicationError`, assert `true`

These tests should initially FAIL since `isRetriable` does not yet handle `*meilisearch.Error`.

## Files

- `internal/indexer/retry_test.go` — modify (add test cases)

## Verification

```bash
go test ./internal/indexer/ -run "TestIsRetriable_MeiliSearch" -v
```

Expect: all 5 tests FAIL (Red phase).
