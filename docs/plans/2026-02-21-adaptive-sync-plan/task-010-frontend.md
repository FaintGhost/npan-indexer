# Task 010: Frontend update

**depends-on**: task-001

## Description

Update the frontend Zod schemas and sync progress display component to show the sync mode and incremental-specific stats.

## Execution Context

**Task Number**: 010 of 010
**Phase**: Refinement
**Prerequisites**: Task 001 (model types defined — frontend schema mirrors backend)

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 17 (incremental mode display), 18 (full mode stats)

## Files to Modify/Create

- Modify: `web/src/lib/sync-schemas.ts`
- Modify: `web/src/components/sync-progress-display.tsx`
- Modify: `web/src/components/sync-progress-display.test.tsx`

## Steps

### Step 1: Update Zod schemas

In `sync-schemas.ts`:
1. Add `IncrementalSyncStatsSchema` with fields: changesFetched, upserted, deleted, skippedUpserts, skippedDeletes, cursorBefore, cursorAfter (all z.number())
2. Add `mode` field to `SyncProgressSchema`: `z.string().optional().default("")`
3. Add `incrementalStats` field: `IncrementalSyncStatsSchema.optional().nullable()`

### Step 2: Update test data

In `sync-progress-display.test.tsx`:
1. Add `mode: "full"` to the `baseProgress` test fixture
2. Create a second test fixture `incrementalProgress` with `mode: "incremental"` and populated `incrementalStats`
3. Add a test that verifies incremental mode renders incremental-specific stats
4. Add a test that verifies full mode renders full-crawl stats

### Step 3: Update SyncProgressDisplay component

In `sync-progress-display.tsx`:
1. Add a mode badge/indicator near the status display (e.g., "全量同步" or "增量同步")
2. When `mode === "incremental"` and `incrementalStats` exists:
   - Show "变更" (changesFetched), "写入" (upserted), "删除" (deleted) stat cards instead of folders/pages
   - Show "跳过写入" and "跳过删除" if skipped counts > 0
3. When `mode === "full"` or mode is empty:
   - Show existing full-crawl stats (folders, pages, files) — no change needed

### Step 4: Verify frontend tests

Run the frontend test suite.

### Step 5: Verify TypeScript compilation

Run the TypeScript type checker.

## Verification Commands

```bash
cd /root/workspace/npan/web && bun test
cd /root/workspace/npan/web && bunx tsc --noEmit
```

## Success Criteria

- Zod schemas include mode and incrementalStats fields
- Component renders mode indicator for both full and incremental
- Incremental mode shows appropriate stat cards
- Full mode rendering is unchanged
- All frontend tests pass
- TypeScript compilation succeeds
