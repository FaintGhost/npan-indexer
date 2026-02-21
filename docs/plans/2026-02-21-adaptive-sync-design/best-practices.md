# Best Practices: Adaptive Sync

## 1. Backward Compatibility

- Old CLI commands (`sync-full`, `sync-incremental`) must continue to work identically.
- Old HTTP endpoints (`/sync/full`, `/sync/full/progress`, `/sync/full/cancel`) must continue to work.
- The progress JSON format must be backward-compatible (new fields are optional).

## 2. Retry Strategy

- **Incremental upserts**: Use `WithRetryVoid` with the same retry policy as full crawl.
- **Incremental deletes**: Use `WithRetryVoid` similarly. Track `SkippedDeletes` on final failure.
- **Fetch changes**: Already uses `WithRetry` in `FetchIncrementalChanges`.

## 3. Progress Persistence

- Incremental sync should persist progress at regular intervals (after each batch of upserts/deletes).
- Use the same `JSONProgressStore` and `SyncProgressState` format.
- Frontend polling works unchanged.

## 4. Verification Consistency

- Both modes call `buildVerification()` with MeiliSearch document count.
- For incremental, `CrawledDocCount` represents the number of successful upserts (not total docs).
- The `Verified` flag and `Warnings` provide the same quality signal.

## 5. Rate Limiting

- Incremental sync should use the same `RequestLimiter` instance as full crawl.
- This ensures consistent rate limiting regardless of mode.
- The `ActivityChecker` integration works for both modes.

## 6. Error Handling

- If incremental fails mid-way, the SyncState cursor is NOT updated (existing behavior preserved).
- The progress store records the error for the frontend to display.
- On context cancellation, progress shows "cancelled" status.

## 7. Testing Strategy

- **Unit tests**: Test mode resolution logic in isolation.
- **Unit tests**: Test `runIncremental()` with mock dependencies.
- **Unit tests**: Test cursor update after successful full crawl.
- **Unit tests**: Test that cursor is NOT updated on failure.
- **Integration**: Verify CLI command parsing and flag forwarding.

## 8. Incremental Batch Processing

- For large change sets, consider batching upserts (e.g., 100 docs per batch).
- Track progress per batch to enable meaningful progress display.
- This matches the full crawl's page-by-page progress pattern.

## 9. Mode Resolution Transparency

- The resolved mode should be logged at sync start.
- The progress state `mode` field allows the frontend to render appropriately.
- CLI output should indicate which mode was auto-selected.
