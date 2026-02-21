# Task 011: Implement reconciliation in sync_manager

**depends-on**: task-009, task-010

## BDD Reference

- Scenario: Successful sync with matching counts
- Scenario: MeiliSearch has fewer documents than crawled
- Scenario: Discovered more files than indexed (some skipped)
- Scenario: Reconciliation failure doesn't fail the sync

## Description

Modify `internal/service/sync_manager.go`:

1. **Add `buildVerification` function** — pure function that takes `meiliCount int64` and `stats models.CrawlStats`, returns `*models.SyncVerification`. Logic:
   - `CrawledDocCount = stats.FilesIndexed + stats.FoldersVisited`
   - `DiscoveredDocCount = stats.FilesDiscovered + stats.FoldersVisited`
   - `SkippedCount = stats.SkippedFiles`
   - `Verified = true`
   - Generate warnings when `meiliCount < CrawledDocCount` or `CrawledDocCount < DiscoveredDocCount`

2. **Call reconciliation in `run()`** — after setting `progress.Status = "done"`, call `m.index.DocumentCount(ctx)`. If it succeeds, call `buildVerification` and set `progress.Verification`. If it fails, leave `Verification` as nil (graceful degradation).

3. **SyncManager needs access to index** — it already has `m.index *search.MeiliIndex`, so `DocumentCount` can be called directly.

## Files

- `internal/service/sync_manager.go` — modify `run()` and add `buildVerification` function

## Verification

```bash
go test ./internal/service/ -run "TestBuildVerification" -v
go build ./...
```

Expect: all reconciliation tests PASS and build succeeds.
