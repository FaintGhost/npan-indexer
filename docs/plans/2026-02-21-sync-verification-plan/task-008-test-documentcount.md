# Task 008: Test DocumentCount method

**depends-on**: none

## BDD Reference

- Scenario: Successful sync with matching counts (needs DocumentCount)
- Scenario: Reconciliation failure doesn't fail the sync (needs DocumentCount to return error)

## Description

Add test in `internal/search/` for the new `DocumentCount` method:

1. `TestMeiliIndex_DocumentCount` — mock `IndexManager` to return `StatsIndex{NumberOfDocuments: 42}`, verify `DocumentCount` returns `42, nil`.

2. `TestMeiliIndex_DocumentCount_Error` — mock returns error, verify `DocumentCount` returns `0, err`.

Use existing test mock patterns from `meili_index_search_test.go` and `meili_index_settings_test.go`.

These tests should fail since `DocumentCount` does not exist yet.

## Files

- `internal/search/meili_index_test.go` — new or modify existing test file

## Verification

```bash
go test ./internal/search/ -run "TestMeiliIndex_DocumentCount" -v
```

Expect: compilation error (Red phase).
