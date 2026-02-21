# Task 008: CLI unification

**depends-on**: task-001

## Description

Add a unified `sync` CLI command that supports `--mode` flag and make the existing `sync-full` and `sync-incremental` commands thin aliases.

## Execution Context

**Task Number**: 008 of 010
**Phase**: Integration
**Prerequisites**: Task 001 (SyncMode type and SyncStartRequest.Mode field exist)

## BDD Scenario Reference

**Spec**: `../2026-02-21-adaptive-sync-design/bdd-specs.md`
**Scenarios**: 12 (sync defaults to auto), 13 (sync-full alias), 14 (sync-incremental alias)

## Files to Modify/Create

- Modify: `internal/cli/root.go`

## Steps

### Step 1: Create newSyncCommand function

Create a new `newSyncCommand(cfg config.Config)` function that:
- Has `Use: "sync"` and `Short: "自适应同步到 Meilisearch"`
- Includes ALL flags from both `sync-full` and `sync-incremental` commands
- Adds a new `--mode` flag with default value "auto" and help text "同步模式: auto|full|incremental"
- Adds incremental-specific flags: `--sync-state-file`, `--window-overlap-ms`, `--incremental-query-words`

### Step 2: Implement sync command RunE

The RunE function should:
1. Parse mode flag
2. Create SyncManager with both full and incremental dependencies
3. Pass `Mode` in `SyncStartRequest`
4. Share the same progress polling loop as `sync-full`
5. The SyncManager handles mode routing internally

### Step 3: Register sync command

In `NewRootCommand`, add `rootCmd.AddCommand(newSyncCommand(cfg))` before the existing sync commands.

### Step 4: Convert sync-full to alias

Modify `newSyncFullCommand` to be a thin wrapper that calls the same logic as `newSyncCommand` but with `Mode: "full"` hardcoded. Alternatively, keep the existing command unchanged since it doesn't pass a Mode field (SyncManager defaults to full when Mode is empty or "auto" with no cursor).

### Step 5: Convert sync-incremental to delegate to SyncManager

Modify `newSyncIncrementalCommand` to create a SyncManager and use `Mode: "incremental"` instead of calling `RunIncrementalSync` directly. This gives incremental the full SyncManager feature set (progress, retry, verification).

### Step 6: Verify

Run the CLI with `--help` to confirm the new `sync` command appears with correct flags. Verify existing `sync-full` and `sync-incremental` commands still work.

## Verification Commands

```bash
cd /root/workspace/npan && go build ./cmd/cli/
cd /root/workspace/npan && ./cmd/cli/main --help
cd /root/workspace/npan && go test ./internal/cli/ -v
```

## Success Criteria

- `npan-cli sync` command exists with --mode flag defaulting to "auto"
- `npan-cli sync-full` still works
- `npan-cli sync-incremental` still works
- All CLI flags are properly wired
- Project compiles successfully
