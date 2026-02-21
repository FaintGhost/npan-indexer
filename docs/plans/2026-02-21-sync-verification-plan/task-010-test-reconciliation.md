# Task 010: Test reconciliation logic

**depends-on**: task-001

## BDD Reference

- Scenario: Successful sync with matching counts
- Scenario: MeiliSearch has fewer documents than crawled
- Scenario: Discovered more files than indexed (some skipped)
- Scenario: Reconciliation failure doesn't fail the sync

## Description

Add tests in `internal/service/` (e.g. `sync_manager_verification_test.go`):

Extract the reconciliation logic into a testable pure function `buildVerification(meiliCount int64, stats models.CrawlStats) *models.SyncVerification` so it can be unit tested without a real MeiliSearch connection.

### Test cases:

1. `TestBuildVerification_MatchingCounts` — FilesIndexed=100, FoldersVisited=20, meiliCount=120. Assert: Verified=true, Warnings empty, MeiliDocCount=120, CrawledDocCount=120.

2. `TestBuildVerification_MeiliFewerThanCrawled` — FilesIndexed=100, FoldersVisited=20, meiliCount=110. Assert: Warnings contains "MeiliSearch 文档数(110) < 爬取写入数(120)".

3. `TestBuildVerification_DiscoveredMoreThanIndexed` — FilesDiscovered=105, FilesIndexed=100, SkippedFiles=5, FoldersVisited=20, meiliCount=120. Assert: Warnings contains gap message, SkippedCount=5.

4. `TestBuildVerification_AllMatch` — FilesDiscovered=100, FilesIndexed=100, SkippedFiles=0, meiliCount=120, FoldersVisited=20. Assert: no warnings.

These tests should fail initially since the function does not exist.

## Files

- `internal/service/sync_manager_verification_test.go` — new file

## Verification

```bash
go test ./internal/service/ -run "TestBuildVerification" -v
```

Expect: compilation error (Red phase).
