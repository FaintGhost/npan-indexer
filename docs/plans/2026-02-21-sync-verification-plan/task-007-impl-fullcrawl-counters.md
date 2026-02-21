# Task 007: Implement counter fix, discovery tracking, and upsert retry+skip in RunFullCrawl

**depends-on**: task-003, task-006

## BDD Reference

- Scenario: FilesIndexed only counts successfully upserted files
- Scenario: FilesDiscovered tracks all files seen from API
- Scenario: Permanent UpsertDocuments failure skips batch
- Scenario: Transient MeiliSearch failure is retried

## Description

Modify `internal/indexer/full_crawl.go` `RunFullCrawl()`:

1. **L2 — Track FilesDiscovered**: After fetching a page, immediately add `stats.FilesDiscovered += int64(len(page.Files))` regardless of upsert outcome.

2. **L3 — Wrap Upsert in WithRetryVoid**: Replace direct `deps.IndexWriter.UpsertDocuments(ctx, docs)` with `WithRetryVoid(ctx, func() error { return deps.IndexWriter.UpsertDocuments(ctx, docs) }, deps.Retry)`.

3. **L1 + L3 — Fix counter and handle failure**:
   - Remove `stats.FilesIndexed += int64(len(page.Files))` from before the upsert
   - On upsert success: `stats.FilesIndexed += filesInBatch`
   - On upsert failure (retry exhausted): `stats.SkippedFiles += filesInBatch`, `stats.FailedRequests++`, continue to next page (do NOT return error)

4. **Preserve existing behavior**: Context cancellation should still propagate (check `ctx.Err()` after failed upsert and return if cancelled).

## Files

- `internal/indexer/full_crawl.go` — modify the upsert section

## Verification

```bash
go test ./internal/indexer/ -run "TestRunFullCrawl_" -v
```

Expect: all 4 tests from Task 006 PASS (Green phase).
