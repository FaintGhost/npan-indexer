# Task 012: Update frontend schemas and tests

**depends-on**: task-001

## BDD Reference

- Scenario: Sync in progress shows discovery stats (schema must support new fields)
- Scenario: Completed sync shows verification result (schema must support verification)

## Description

### Schema changes (`web/src/lib/sync-schemas.ts`):

1. **CrawlStatsSchema** — add:
   - `filesDiscovered: z.number().optional().default(0)`
   - `skippedFiles: z.number().optional().default(0)`

2. **SyncProgressSchema** — add:
   - `verification: z.object({...}).optional().nullable()` with fields: meiliDocCount, crawledDocCount, discoveredDocCount, skippedCount, verified (boolean), warnings (array of string)

### Test changes (`web/src/lib/sync-schemas.test.ts`):

1. Add test: schema validates data with new CrawlStats fields
2. Add test: schema validates data with verification field present
3. Add test: schema validates data without verification (backward compat — old data missing field)
4. Add test: schema validates data with empty verification warnings

### Test changes (`web/src/components/sync-progress-display.test.tsx`):

Update `baseProgress` fixture to include `filesDiscovered` and `skippedFiles` in `aggregateStats`.

## Files

- `web/src/lib/sync-schemas.ts` — modify
- `web/src/lib/sync-schemas.test.ts` — modify
- `web/src/components/sync-progress-display.test.tsx` — modify fixture

## Verification

```bash
cd web && bun run typecheck && bun run test -- --run
```

Expect: all tests pass, typecheck clean.
