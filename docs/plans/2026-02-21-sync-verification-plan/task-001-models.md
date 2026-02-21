# Task 001: Add new fields to CrawlStats and SyncProgressState models

**depends-on**: none

## BDD Reference

Supports all scenarios — this task provides the data structures that all other tasks depend on.

## Description

Modify `internal/models/models.go` to add:

1. **CrawlStats** — add two new fields:
   - `FilesDiscovered int64` with JSON tag `"filesDiscovered"`
   - `SkippedFiles int64` with JSON tag `"skippedFiles"`

2. **SyncVerification** — new struct:
   - `MeiliDocCount int64` — document count from MeiliSearch
   - `CrawledDocCount int64` — FilesIndexed + FoldersVisited
   - `DiscoveredDocCount int64` — FilesDiscovered + FoldersVisited
   - `SkippedCount int64`
   - `Verified bool`
   - `Warnings []string` with `omitempty`

3. **SyncProgressState** — add:
   - `Verification *SyncVerification` with JSON tag `"verification,omitempty"`

## Files

- `internal/models/models.go` — modify

## Verification

```bash
go build ./...
```

Ensure compilation passes. No test changes needed since these are additive struct fields with zero-value defaults.
