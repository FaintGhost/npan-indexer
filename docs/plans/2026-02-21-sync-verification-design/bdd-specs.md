# BDD Specifications

## Feature: Accurate File Indexing Counter

### Scenario: FilesIndexed only counts successfully upserted files
```gherkin
Given a folder with 10 files on page 0
When UpsertDocuments succeeds for that page
Then stats.FilesIndexed should be 10

Given a folder with 10 files on page 0
When UpsertDocuments fails for that page
Then stats.FilesIndexed should be 0
And stats.SkippedFiles should be 10
And stats.FailedRequests should be 1
```

### Scenario: FilesDiscovered tracks all files seen from API
```gherkin
Given a folder with 2 pages, 5 files each
When both pages are fetched successfully
Then stats.FilesDiscovered should be 10
Regardless of whether UpsertDocuments succeeds or fails
```

## Feature: Upsert Retry and Skip

### Scenario: Transient MeiliSearch failure is retried
```gherkin
Given RetryPolicyOptions with MaxRetries=3
When UpsertDocuments fails with HTTP 503 on first attempt
And succeeds on second attempt
Then stats.FilesIndexed should include those files
And stats.FailedRequests should be 0
And crawl should continue to next page
```

### Scenario: Permanent UpsertDocuments failure skips batch
```gherkin
Given RetryPolicyOptions with MaxRetries=3
When UpsertDocuments fails with HTTP 400 (not retriable)
Then stats.SkippedFiles should include those files
And stats.FailedRequests should be 1
And crawl should continue to next page (NOT terminate)
```

### Scenario: MeiliSearch timeout is retriable
```gherkin
Given a *meilisearch.Error with ErrCode=MeilisearchTimeoutError
When isRetriable is called
Then it should return true
```

### Scenario: MeiliSearch 429 is retriable
```gherkin
Given a *meilisearch.Error with StatusCode=429
When isRetriable is called
Then it should return true
```

## Feature: Post-Sync Reconciliation

### Scenario: Successful sync with matching counts
```gherkin
Given a completed sync with FilesIndexed=100, FoldersVisited=20
And MeiliSearch DocumentCount returns 120
Then verification.Verified should be true
And verification.Warnings should be empty
And verification.MeiliDocCount should be 120
And verification.CrawledDocCount should be 120
```

### Scenario: MeiliSearch has fewer documents than crawled
```gherkin
Given a completed sync with FilesIndexed=100, FoldersVisited=20
And MeiliSearch DocumentCount returns 110
Then verification.Verified should be true
And verification.Warnings should contain "MeiliSearch 文档数(110) < 爬取写入数(120)"
```

### Scenario: Discovered more files than indexed (some skipped)
```gherkin
Given a completed sync with FilesDiscovered=105, FilesIndexed=100, SkippedFiles=5
Then verification.Warnings should contain information about the gap
And verification.SkippedCount should be 5
```

### Scenario: Reconciliation failure doesn't fail the sync
```gherkin
Given a completed sync
When MeiliSearch GetStats returns an error
Then progress.Status should still be "done"
And progress.Verification should be nil (graceful degradation)
```

## Feature: Frontend Verification Display

### Scenario: Sync in progress shows discovery stats
```gherkin
Given a running sync with FilesDiscovered=50, FilesIndexed=30
Then the progress display should show both counters
```

### Scenario: Completed sync shows verification result
```gherkin
Given a completed sync with verification.Verified=true and no warnings
Then the UI should show a green checkmark with "验证通过"

Given a completed sync with verification warnings
Then the UI should show a yellow warning banner with warning messages
```

## Testing Strategy

### Unit Tests (Go)

| Test | File | What it validates |
|------|------|------------------|
| `TestFilesIndexedAfterUpsert` | `full_crawl_test.go` | Counter increments only after successful upsert |
| `TestFilesDiscoveredPerPage` | `full_crawl_test.go` | Discovery counter tracks all API-returned files |
| `TestSkippedFilesOnUpsertFailure` | `full_crawl_test.go` | SkippedFiles counts failed batches, crawl continues |
| `TestWithRetryVoid` | `retry_test.go` | Void wrapper delegates correctly to WithRetry |
| `TestIsRetriableMeiliError` | `retry_test.go` | MeiliSearch errors correctly classified |
| `TestReconciliation` | `sync_manager_test.go` | Verification struct populated correctly |
| `TestReconciliationGracefulDegradation` | `sync_manager_test.go` | GetStats failure doesn't break sync |

### Frontend Tests (Vitest)

| Test | File | What it validates |
|------|------|------------------|
| `shows filesDiscovered stat` | `sync-progress-display.test.tsx` | New stat card renders |
| `shows verification success` | `sync-progress-display.test.tsx` | Green check on no warnings |
| `shows verification warnings` | `sync-progress-display.test.tsx` | Yellow banner on warnings |
| `schema accepts verification field` | `sync-schemas.test.ts` | Zod schema validates new fields |
