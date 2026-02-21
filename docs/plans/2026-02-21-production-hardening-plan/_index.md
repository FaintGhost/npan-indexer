# npan 生产化加固实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 将 npan 从原型阶段全面升级为生产级服务，消除 13 项安全审计发现，实现双路径认证、路由重设计、运维增强。

**Architecture:** 基于 Echo v5 中间件链实现分层安全架构。公开端点 (`/api/v1/app/*`) 使用 EmbeddedAuth 自动注入服务端凭据；外部 API (`/api/v1/*`) 和管理端点 (`/api/v1/admin/*`) 要求 X-API-Key 认证。所有安全能力（速率限制、CORS、body limit、安全头）通过中间件实现。

**Tech Stack:** Go 1.25, Echo v5, Meilisearch, golang.org/x/time/rate, crypto/subtle

**Design Support:**
- [BDD Specs](../2026-02-21-production-hardening-design/bdd-specs.md)
- [Architecture](../2026-02-21-production-hardening-design/architecture.md)
- [Best Practices](../2026-02-21-production-hardening-design/best-practices.md)
- [Requirements](../2026-02-21-production-hardening-design/_index.md)

## Execution Plan

### Phase 1: Foundation

- [Task 001: Create unified error response types](./task-001-unified-error-response.md)

### Phase 2: Config Validation

- [Task 002: Test config startup validation](./task-002-test-config-validation.md)
- [Task 003: Implement config startup validation](./task-003-impl-config-validation.md)

### Phase 3: Authentication

- [Task 004: Test API Key auth middleware](./task-004-test-apikey-auth.md)
- [Task 005: Implement API Key auth middleware](./task-005-impl-apikey-auth.md)
- [Task 006: Test embedded auth middleware](./task-006-test-embedded-auth.md)
- [Task 007: Implement embedded auth middleware](./task-007-impl-embedded-auth.md)

### Phase 4: Route Restructure

- [Task 008: Test route structure with auth enforcement](./task-008-test-route-auth-enforcement.md)
- [Task 009: Implement route restructure](./task-009-impl-route-restructure.md)

### Phase 5: Input Validation

- [Task 010: Test pageSize validation](./task-010-test-pagesize-validation.md)
- [Task 011: Implement pageSize validation](./task-011-impl-pagesize-validation.md)
- [Task 012: Test type parameter whitelist](./task-012-test-type-whitelist.md)
- [Task 013: Implement type parameter whitelist](./task-013-impl-type-whitelist.md)
- [Task 014: Test checkpoint path validation](./task-014-test-checkpoint-path.md)
- [Task 015: Implement checkpoint path validation](./task-015-impl-checkpoint-path.md)

### Phase 6: Error Handling

- [Task 016: Test error response sanitization](./task-016-test-error-sanitization.md)
- [Task 017: Implement error response sanitization](./task-017-impl-error-sanitization.md)
- [Task 018: Test progress response DTO](./task-018-test-progress-dto.md)
- [Task 019: Implement progress response DTO](./task-019-impl-progress-dto.md)

### Phase 7: Security Middleware

- [Task 020: Test rate limiting](./task-020-test-rate-limiting.md)
- [Task 021: Implement rate limiting](./task-021-impl-rate-limiting.md)
- [Task 022: Implement security middleware stack](./task-022-impl-security-middleware-stack.md)

### Phase 8: Credential Security

- [Task 023: Test config log sanitization](./task-023-test-config-log-sanitization.md)
- [Task 024: Implement credential security](./task-024-impl-credential-security.md)

### Phase 9: Health & Ops

- [Task 025: Test health check endpoints](./task-025-test-health-endpoints.md)
- [Task 026: Implement health check endpoints](./task-026-impl-health-endpoints.md)
- [Task 027: Implement graceful shutdown](./task-027-impl-graceful-shutdown.md)
- [Task 028: Implement io.LimitReader](./task-028-impl-io-limit-reader.md)

### Phase 10: Frontend & Deployment

- [Task 029: Implement frontend migration](./task-029-impl-frontend-migration.md)
- [Task 030: Implement Dockerfile](./task-030-impl-dockerfile.md)

### Phase 11: Cleanup & Verification

- [Task 031: Git credential cleanup](./task-031-git-credential-cleanup.md)
- [Task 032: Integration verification](./task-032-integration-verification.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-21-production-hardening-plan/`. Execution options:

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
