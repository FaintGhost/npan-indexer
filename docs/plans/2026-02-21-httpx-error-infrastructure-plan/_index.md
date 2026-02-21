# HTTP 错误响应基础设施实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 在 `internal/httpx/` 包中创建统一的错误响应基础设施，包括 `errors.go` 和 `errors_test.go`。

**Architecture:** 新增 `ErrorResponse` struct 及错误码常量，提供 `writeErrorResponse` 函数替代现有 `writeError`，并注册 `customHTTPErrorHandler` 作为 Echo 全局错误处理器，实现结构化 JSON 错误响应和 slog 日志记录。

**Tech Stack:** Go 1.25, Echo v5 (`github.com/labstack/echo/v5`), `log/slog`, `net/http/httptest`

**Execution Plan:**
- [Task 001: 编写错误响应测试 (Red)](./task-001-write-error-response-tests.md)
- [Task 002: 实现 errors.go (Green)](./task-002-implement-errors.md)

---

## Execution Handoff

**Plan complete and saved to `docs/plans/2026-02-21-httpx-error-infrastructure-plan/`. Execution options:**

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
