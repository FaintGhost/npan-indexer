# Adaptive Sync Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** Unify sync-full and sync-incremental into a single adaptive sync interface that auto-detects the correct mode.

**Architecture:** SyncManager becomes the unified orchestrator. A `resolveMode` function determines full vs incremental based on SyncState cursor. A new `runIncremental` method handles incremental logic within SyncManager with the same quality features (retry, rate limiting, progress, verification).

**Tech Stack:** Go 1.24, MeiliSearch, React 19 + TypeScript + Zod

**Design Support:**
- [BDD Specs](../2026-02-21-adaptive-sync-design/bdd-specs.md)
- [Architecture](../2026-02-21-adaptive-sync-design/architecture.md)
- [Best Practices](../2026-02-21-adaptive-sync-design/best-practices.md)

**Execution Plan:**
- [Task 001: Add model types and SyncManager dependency changes](./task-001-models-foundation.md)
- [Task 002: RED - Mode resolution tests](./task-002-red-resolve-mode.md)
- [Task 003: RED - runIncremental tests](./task-003-red-run-incremental.md)
- [Task 004: RED - Mode routing and cursor update tests](./task-004-red-mode-routing-cursor.md)
- [Task 005: GREEN - resolveMode implementation](./task-005-green-resolve-mode.md)
- [Task 006: GREEN - runIncremental implementation](./task-006-green-run-incremental.md)
- [Task 007: GREEN - Mode routing and cursor update](./task-007-green-mode-routing-cursor.md)
- [Task 008: CLI unification](./task-008-cli-unification.md)
- [Task 009: HTTP API update](./task-009-http-api.md)
- [Task 010: Frontend update](./task-010-frontend.md)

---

## Dependency Graph

```
Tier 0:  [001]
          │
Tier 1:  [002] [003] [004] [008] [009] [010]
          │      │      │
Tier 2:  [005] [006]   │
          │      │      │
Tier 3:  └──────┴──[007]
```

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-21-adaptive-sync-plan/`. Execution options:

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
