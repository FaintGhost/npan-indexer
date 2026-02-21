# Prometheus Metrics Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** Add Prometheus metrics to npan backend, covering HTTP requests, sync tasks, Meilisearch/search operations, and Go runtime, served on an independent port.

**Architecture:** Use `echo-contrib/echoprometheus` for HTTP request metrics and `prometheus/client_golang` for custom business metrics. All metrics registered in a custom `prometheus.Registry` (not global). Instrumentation via decorator pattern (consistent with existing `CachedQueryService` pattern). Metrics served on a separate HTTP server on port `:9091`.

**Tech Stack:** `prometheus/client_golang`, `echo-contrib/echoprometheus`, `promhttp`, Go `net/http`

**Design Support:**
- [BDD Specs](../2026-02-22-prometheus-metrics-design/bdd-specs.md)
- [Architecture](../2026-02-22-prometheus-metrics-design/architecture.md)
- [Best Practices](../2026-02-22-prometheus-metrics-design/best-practices.md)

## Execution Plan

- [Task 001: Add dependencies and config](./task-001-add-dependencies-and-config.md)
- [Task 002: Create metrics registry with Go runtime collectors](./task-002-create-metrics-registry.md)
- [Task 003: Create SyncMetrics definitions](./task-003-create-sync-metrics.md)
- [Task 004: Create SearchMetrics definitions](./task-004-create-search-metrics.md)
- [Task 005: Create metrics server](./task-005-create-metrics-server.md)
- [Task 006: Create SyncReporter interface and implementation](./task-006-create-sync-reporter.md)
- [Task 007: Extract IndexOperator interface and create MeiliIndex instrumenter](./task-007-create-meili-instrumenter.md)
- [Task 008: Create search service instrumenter](./task-008-create-search-instrumenter.md)
- [Task 009: Integrate echoprometheus middleware](./task-009-integrate-echoprometheus.md)
- [Task 010: Integrate SyncReporter into SyncManager](./task-010-integrate-sync-reporter.md)
- [Task 011: Wire metrics in main.go with graceful shutdown](./task-011-wire-main-graceful-shutdown.md)
- [Task 012: Update Dockerfile and docker-compose](./task-012-update-infra.md)

## Dependency Graph

```
001 (deps+config)
├── 002 (registry)
│   ├── 003 (sync_metrics)  ──── 006 (sync_reporter) ──── 010 (SyncManager integration)
│   ├── 004 (search_metrics)
│   │   ├── 007 (meili_instrumenter)
│   │   └── 008 (search_instrumenter)
│   └── 005 (metrics_server)
└── 009 (echoprometheus) ← independent of 003-008

011 (main.go wiring) ← depends on 005,006,007,008,009,010
└── 012 (infra) ← depends on 011
```

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-22-prometheus-metrics-plan/`. Execution options:

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
