# Task 009: Implement DocumentCount method

**depends-on**: task-008

## BDD Reference

- Scenario: Successful sync with matching counts
- Scenario: Reconciliation failure doesn't fail the sync

## Description

Add `DocumentCount` method to `internal/search/meili_index.go`:

- Signature: `func (m *MeiliIndex) DocumentCount(ctx context.Context) (int64, error)`
- Call `m.index.GetStatsWithContext(ctx)`
- Return `stats.NumberOfDocuments` on success, `0, err` on failure

Also update any test mock structs (in search test files) that implement `IndexManager` to include the `GetStatsWithContext` method if not already present.

## Files

- `internal/search/meili_index.go` — modify (add method)

## Verification

```bash
go test ./internal/search/ -run "TestMeiliIndex_DocumentCount" -v
```

Expect: tests PASS (Green phase).
