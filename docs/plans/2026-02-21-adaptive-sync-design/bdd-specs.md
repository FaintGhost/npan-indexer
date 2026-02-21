# BDD Specifications: Adaptive Sync

## Feature: Sync Mode Auto-Detection

### Scenario 1: First run triggers full crawl
```gherkin
Given no SyncState file exists
When a sync is started with mode "auto"
Then the system should execute a full crawl
And the progress state should show mode "full"
```

### Scenario 2: Existing cursor triggers incremental
```gherkin
Given a SyncState file exists with LastSyncTime > 0
When a sync is started with mode "auto"
Then the system should execute an incremental sync
And the progress state should show mode "incremental"
```

### Scenario 3: Explicit full mode overrides auto
```gherkin
Given a SyncState file exists with LastSyncTime > 0
When a sync is started with mode "full"
Then the system should execute a full crawl regardless of cursor state
```

### Scenario 4: Explicit incremental mode with no cursor
```gherkin
Given no SyncState file exists
When a sync is started with mode "incremental"
Then the system should execute an incremental sync with since=0
```

## Feature: Post-Full-Crawl Cursor Update

### Scenario 5: Full crawl updates SyncState cursor
```gherkin
Given a full crawl completes successfully
When the run finishes
Then SyncState.LastSyncTime should be set to the crawl end time
And the next auto-detected sync should run in incremental mode
```

### Scenario 6: Failed full crawl does not update cursor
```gherkin
Given a full crawl fails with an error
When the run finishes
Then SyncState.LastSyncTime should NOT be updated
```

## Feature: Incremental Sync in SyncManager

### Scenario 7: Incremental sync with retry on upsert failure
```gherkin
Given an incremental sync is running
When an upsert batch fails with a retriable error
Then the system should retry up to MaxRetries times
And track skipped upserts if all retries fail
```

### Scenario 8: Incremental sync with retry on delete failure
```gherkin
Given an incremental sync is running
When a delete batch fails with a retriable error
Then the system should retry up to MaxRetries times
And track skipped deletes if all retries fail
```

### Scenario 9: Incremental sync progress tracking
```gherkin
Given an incremental sync is running
When changes are being processed
Then the progress state should show incremental stats (changesFetched, upserted, deleted)
And the progress state should be persisted to the progress file
```

### Scenario 10: Incremental sync verification
```gherkin
Given an incremental sync completes successfully
When the run finishes
Then buildVerification should be called with MeiliSearch document count
And the result should be stored in progress.Verification
```

## Feature: Incremental Rate Limiting

### Scenario 11: Incremental sync uses shared rate limiter
```gherkin
Given an incremental sync is running
And the RequestLimiter is configured with maxConcurrent=2 and minTimeMS=200
When upsert and delete operations are submitted
Then they should respect the rate limiter's concurrency and timing constraints
```

## Feature: Unified CLI

### Scenario 12: sync command defaults to auto mode
```gherkin
Given the user runs "npan-cli sync"
When no --mode flag is provided
Then the default mode should be "auto"
```

### Scenario 13: sync-full is aliased to sync --mode full
```gherkin
Given the user runs "npan-cli sync-full"
Then the behavior should be identical to "npan-cli sync --mode full"
```

### Scenario 14: sync-incremental is aliased to sync --mode incremental
```gherkin
Given the user runs "npan-cli sync-incremental"
Then the behavior should be identical to "npan-cli sync --mode incremental"
```

## Feature: Unified HTTP API

### Scenario 15: POST /api/v1/admin/sync/start accepts mode field
```gherkin
Given the admin API is available
When a POST request is sent to /api/v1/admin/sync/start with {"mode": "auto"}
Then the SyncManager should start with the specified mode
And the response should indicate the sync has started
```

### Scenario 16: GET /api/v1/admin/sync/progress returns mode
```gherkin
Given a sync is in progress
When GET /api/v1/admin/sync/progress is called
Then the response should include the "mode" field indicating "full" or "incremental"
```

## Feature: Frontend Mode Display

### Scenario 17: Progress display shows sync mode
```gherkin
Given a sync is in progress with mode "incremental"
When the progress display renders
Then it should show "incremental" mode indicator
And show incremental-specific stats (changes, upserts, deletes)
```

### Scenario 18: Progress display shows full mode stats
```gherkin
Given a sync is in progress with mode "full"
When the progress display renders
Then it should show "full" mode indicator
And show full-crawl-specific stats (folders, pages, files)
```
