# Playwright E2E Testing Implementation Plan

## Goal

为 npan 项目添加 Playwright E2E 测试，覆盖搜索、下载、管理后台全流程及边界场景，集成到 Docker Compose CI 和 GitHub Actions。

## Design Reference

- [Design Document](../2026-02-22-playwright-e2e-design/_index.md)
- [BDD Specifications](../2026-02-22-playwright-e2e-design/bdd-specs.md)
- [Architecture](../2026-02-22-playwright-e2e-design/architecture.md)
- [Best Practices](../2026-02-22-playwright-e2e-design/best-practices.md)

## Constraints

- 使用 `bun` 作为包管理器（本地开发）
- Playwright 官方 Docker 镜像使用 `npx`（容器内无 bun）
- 单 worker，仅 Chromium，避免 Meilisearch 数据竞争
- 下载 API 必须 mock（CI 中 NPA_TOKEN 为 dummy）
- `ipc: host` + `init: true` 防止 Chromium 容器崩溃

## Dependency Graph

```
Task 001 (Playwright 项目基础设施) ──┬──→ Task 003 (搜索测试) ──→ Task 005 (下载测试)
                                     │                        └──→ Task 007 (边界测试)
                                     └──→ Task 004 (Admin 认证测试) ──→ Task 006 (Admin 同步测试)

Task 002 (Docker CI 集成)

Task 008 (全链路验证) ← depends on ALL above
```

## Execution Plan

- [Task 001: Playwright 项目基础设施](./task-001-playwright-infra.md)
- [Task 002: Docker Compose CI 集成](./task-002-docker-ci.md)
- [Task 003: 搜索流程 E2E 测试](./task-003-search-tests.md)
- [Task 004: Admin 认证 E2E 测试](./task-004-admin-auth-tests.md)
- [Task 005: 下载流程 E2E 测试](./task-005-download-tests.md)
- [Task 006: Admin 同步 E2E 测试](./task-006-admin-sync-tests.md)
- [Task 007: 边界场景 E2E 测试](./task-007-edge-case-tests.md)
- [Task 008: 全链路验证](./task-008-full-verification.md)

## Commit Boundaries

| Commit | Tasks | 描述 |
|--------|-------|------|
| 1 | 001 + 002 | `feat: add playwright e2e infrastructure and CI integration` |
| 2 | 003 + 004 + 005 | `test(e2e): search, download, and admin auth tests` |
| 3 | 006 + 007 | `test(e2e): admin sync and edge case tests` |
| 4 | 008 | `ci: verify full e2e pipeline` (only if CI config adjustments needed) |
