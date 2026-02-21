# 2C2G 性能加固实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 针对 2C2G 同机部署（npan-server + Meilisearch）进行性能加固，解决 HTTP 超时缺失、限流未挂载、搜索无缓存、同步无动态降速等问题。

**Architecture:** 新增 CachedQueryService 装饰器包装 QueryService 提供 LRU+TTL 缓存；新增 SearchActivityTracker 追踪搜索活动；修改 RequestLimiter 支持动态速率调整；在 NewServer 中挂载已有的 RateLimitMiddleware；在 http.Server 中配置超时参数。

**Tech Stack:** Go 1.25, Echo v5, hashicorp/golang-lru/v2/expirable, golang.org/x/time/rate

**Design Support:**
- [BDD Specs](../2026-02-21-performance-hardening-design/bdd-specs.md)
- [Architecture](../2026-02-21-performance-hardening-design/architecture.md)
- [Best Practices](../2026-02-21-performance-hardening-design/best-practices.md)
- [Requirements](../2026-02-21-performance-hardening-design/_index.md)

## Execution Plan

### Phase 1: HTTP Server Timeouts

- [Task 001: Test config timeout fields](./task-001-test-config-timeouts.md)
- [Task 002: Implement config timeouts and http.Server wiring](./task-002-impl-config-timeouts.md)

### Phase 2: Rate Limit Mounting

- [Task 003: Test rate limit middleware mounting](./task-003-test-ratelimit-mounting.md)
- [Task 004: Mount rate limit middleware](./task-004-impl-ratelimit-mounting.md)

### Phase 3: Search Cache

- [Task 005: Test CachedQueryService](./task-005-test-cached-query-service.md)
- [Task 006: Implement CachedQueryService](./task-006-impl-cached-query-service.md)

### Phase 4: Activity Tracking

- [Task 007: Test SearchActivityTracker](./task-007-test-activity-tracker.md)
- [Task 008: Implement SearchActivityTracker](./task-008-impl-activity-tracker.md)

### Phase 5: Sync Dynamic Throttle

- [Task 009: Test dynamic throttle in RequestLimiter](./task-009-test-dynamic-throttle.md)
- [Task 010: Implement dynamic throttle](./task-010-impl-dynamic-throttle.md)

### Phase 6: Integration & Runtime

- [Task 011: Integration wiring in main.go](./task-011-impl-integration-wiring.md)
- [Task 012: Dockerfile GOMEMLIMIT](./task-012-impl-dockerfile-gomemlimit.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-21-performance-hardening-plan/`. Execution options:

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
