# Task 014: Build verification and final commit

**depends-on**: task-007, task-011, task-013

## Description

Run full build and test suite across both Go and frontend to verify everything integrates correctly.

## Verification

```bash
# Go
go build ./...
go test ./...

# Frontend
cd web
bun run typecheck
bun run test -- --run
bun run build
```

All must pass. Then commit all changes.

## Commit

Single commit with message:

```
feat(sync): add 4-layer verification for file indexing completeness

- L1: Fix FilesIndexed counter to increment only after successful upsert
- L2: Track FilesDiscovered for crawl-time discovered vs indexed comparison
- L3: Add Upsert retry with WithRetryVoid and skip-on-failure behavior
- L4: Post-sync reconciliation comparing MeiliSearch document count vs crawl stats
- Frontend: Display discovery stats and verification results in admin panel
```
