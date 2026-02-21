# Task 003: Implement isRetriable MeiliSearch support

**depends-on**: task-002

## BDD Reference

- Scenario: MeiliSearch timeout is retriable
- Scenario: MeiliSearch 429 is retriable

## Description

Modify `isRetriable()` in `internal/indexer/retry.go` to recognize `*meilisearch.Error`:

1. Import `github.com/meilisearch/meilisearch-go`
2. After the existing `*npan.StatusError` check, add a new block using `errors.As(err, &meiliErr)`
3. For `MeilisearchTimeoutError` and `MeilisearchCommunicationError` → return `true`
4. For `MeilisearchApiError` and `MeilisearchApiErrorWithoutMessage` → return `true` only if `StatusCode == 429` or `StatusCode >= 500`
5. All other MeiliSearch error codes → return `false`

## Files

- `internal/indexer/retry.go` — modify `isRetriable` function

## Verification

```bash
go test ./internal/indexer/ -run "TestIsRetriable_MeiliSearch" -v
```

Expect: all 5 tests PASS (Green phase).
