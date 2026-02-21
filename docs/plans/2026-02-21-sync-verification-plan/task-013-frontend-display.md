# Task 013: Update sync progress display with verification UI

**depends-on**: task-012

## BDD Reference

- Scenario: Sync in progress shows discovery stats
- Scenario: Completed sync shows verification result

## Description

### Display changes (`web/src/components/sync-progress-display.tsx`):

1. **Stats cards** — add "已发现" card showing `aggregateStats.filesDiscovered`. If `skippedFiles > 0`, add "已跳过" card with rose color.

2. **Verification section** — shown only when `progress.verification` is not null:
   - If `verification.warnings` is empty: green checkmark icon with "验证通过" text, plus summary line showing meiliDocCount, crawledDocCount, discoveredDocCount
   - If `verification.warnings` is non-empty: yellow warning banner listing each warning message

### Test changes (`web/src/components/sync-progress-display.test.tsx`):

1. `shows filesDiscovered stat` — render with `filesDiscovered: 50`, assert "已发现" text and value present
2. `shows verification success` — render with verification object (no warnings), assert "验证通过" text
3. `shows verification warnings` — render with verification object containing warnings, assert warning text visible
4. `hides verification when null` — render without verification, assert no verification section

## Files

- `web/src/components/sync-progress-display.tsx` — modify
- `web/src/components/sync-progress-display.test.tsx` — modify

## Verification

```bash
cd web && bun run test -- --run
```

Expect: all tests pass including new verification display tests.
