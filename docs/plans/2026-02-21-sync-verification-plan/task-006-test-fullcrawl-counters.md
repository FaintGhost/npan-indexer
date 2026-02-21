# Task 006: Test accurate counters and skip behavior in RunFullCrawl

**depends-on**: task-001, task-005

## BDD Reference

- Scenario: FilesIndexed only counts successfully upserted files
- Scenario: FilesDiscovered tracks all files seen from API
- Scenario: Permanent UpsertDocuments failure skips batch
- Scenario: Transient MeiliSearch failure is retried

## Description

Create `internal/indexer/full_crawl_test.go` with test doubles:

### Test doubles needed:

- `mockAPI` implementing `npan.API` — `ListFolderChildren` returns configurable pages
- `mockIndexWriter` implementing `IndexWriter` — tracks calls, can be configured to fail on specific calls
- `mockCheckpointStore` implementing `CheckpointStore` — in-memory save/load/clear

### Test cases:

1. `TestRunFullCrawl_FilesIndexedAfterUpsert` — 1 folder, 1 page, 10 files. Upsert succeeds. Assert `stats.FilesIndexed == 10` and `stats.FilesDiscovered == 10`.

2. `TestRunFullCrawl_FilesDiscoveredPerPage` — 1 folder, 2 pages, 5 files each. All upserts succeed. Assert `stats.FilesDiscovered == 10`.

3. `TestRunFullCrawl_SkippedFilesOnUpsertFailure` — 1 folder, 2 pages. First page upsert fails (non-retriable). Second page succeeds. Assert:
   - `stats.FilesDiscovered == 10`
   - `stats.FilesIndexed == 5` (only second page)
   - `stats.SkippedFiles == 5` (first page)
   - `stats.FailedRequests == 1`
   - Crawl completes (does NOT terminate on first page failure)

4. `TestRunFullCrawl_UpsertRetrySuccess` — Upsert fails with retriable error on first attempt, succeeds on second. Assert `stats.FilesIndexed` includes those files and `stats.SkippedFiles == 0`.

These tests should initially FAIL since the implementation has not been changed yet.

## Files

- `internal/indexer/full_crawl_test.go` — new file

## Verification

```bash
go test ./internal/indexer/ -run "TestRunFullCrawl_" -v
```

Expect: tests FAIL (Red phase) — counters are wrong and crawl terminates on upsert failure.
