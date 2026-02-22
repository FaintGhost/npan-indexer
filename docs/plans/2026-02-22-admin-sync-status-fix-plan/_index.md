# Admin 同步状态不自动刷新 Bug 修复计划

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 修复 /admin 页面点击"启动同步"后 UI 不自动刷新为 running 状态的 bug。

**Architecture:** 后端 `GetProgress()` 存在竞态条件：`Start()` 方法通过 goroutine 异步执行同步，但在 goroutine 写入初始 "running" 进度之前，`GetProgress()` 返回旧的文件存储数据。前端轮询收到非 running 状态后立即停止轮询，导致 UI 永远无法更新。修复需要同时调整后端状态一致性和前端轮询鲁棒性。

**Tech Stack:** Go (后端)、React/TypeScript (前端)、Playwright (E2E 测试)

**Root Cause:**
1. 后端 `SyncManager.Start()` 在 goroutine 启动后立即返回，但 goroutine 需要时间发现根目录等初始化工作，此时 progress store 里还是旧数据
2. 后端 `GetProgress()` 当 `IsRunning()=true` 但 progress store 非 running 时，未纠正状态
3. 前端 `startPolling` 在首次收到非 running 状态后立即停止轮询

**Execution Plan:**
- [Task 001: 后端 GetProgress 竞态测试 (Red)](./task-001-red-get-progress-race.md)
- [Task 002: 后端 GetProgress 竞态修复 (Green)](./task-002-green-get-progress-race.md)
- [Task 003: 前端轮询鲁棒性测试 (Red)](./task-003-red-frontend-polling.md)
- [Task 004: 前端轮询鲁棒性实现 (Green)](./task-004-green-frontend-polling.md)
- [Task 005: E2E 同步状态自动刷新测试](./task-005-e2e-sync-status-refresh.md)

---

## Execution Handoff

**Plan complete and saved to `docs/plans/2026-02-22-admin-sync-status-fix-plan/`. Execution options:**

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
