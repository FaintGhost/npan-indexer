# Npan React 前端重写实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 将 Npan 的单文件 HTML 前端重写为 React 19 + TypeScript SPA，保留所有现有功能并新增 Admin 管理页面。

**Architecture:** Vite 7 构建的 React 19 SPA，使用 TanStack Router 实现文件路由（搜索页 `/app` + 管理页 `/app/admin`），Zod 校验所有 API 响应，Tailwind CSS v4 CSS-first 配置。构建产物输出到 `web/dist/`，由 Go `embed.FS` 嵌入服务。

**Tech Stack:** React 19, Vite 7, TypeScript (strict), Tailwind CSS v4, TanStack Router, Zod, oxlint/oxfmt, Vitest + React Testing Library + MSW

**Design Support:**
- [BDD Specs](../2026-02-21-react-frontend-rewrite-design/bdd-specs.md)
- [Best Practices](../2026-02-21-react-frontend-rewrite-design/best-practices.md)
- [Design Overview](../2026-02-21-react-frontend-rewrite-design/_index.md)

**Execution Plan:**

### Phase 1: 项目基础设施

- [Task 001: 初始化 Vite + React 19 + TypeScript 项目](./task-001-init-vite-project.md)
- [Task 002: 配置 Tailwind CSS v4 + 设计 token](./task-002-setup-tailwind-v4.md)
- [Task 003: 配置 TanStack Router + 文件路由](./task-003-setup-tanstack-router.md)
- [Task 004: 配置 Vitest + React Testing Library + MSW](./task-004-setup-testing.md)

### Phase 2: Zod Schema 与 API 客户端

- [Task 005: 测试 Zod Schema（搜索/下载/错误响应）](./task-005-red-zod-schemas.md)
- [Task 006: 实现 Zod Schema（搜索/下载/错误响应）](./task-006-green-zod-schemas.md)
- [Task 007: 测试 Zod Schema（同步进度）](./task-007-red-zod-sync-schema.md)
- [Task 008: 实现 Zod Schema（同步进度）](./task-008-green-zod-sync-schema.md)
- [Task 009: 测试 API 客户端封装](./task-009-red-api-client.md)
- [Task 010: 实现 API 客户端封装](./task-010-green-api-client.md)

### Phase 3: 工具函数与基础组件

- [Task 011: 测试工具函数（formatBytes, formatTime, getFileIcon）](./task-011-red-utils.md)
- [Task 012: 实现工具函数](./task-012-green-utils.md)
- [Task 013: 测试骨架屏与空状态组件](./task-013-red-empty-states.md)
- [Task 014: 实现骨架屏与空状态组件](./task-014-green-empty-states.md)
- [Task 015: 测试文件卡片组件](./task-015-red-file-card.md)
- [Task 016: 实现文件卡片组件](./task-016-green-file-card.md)

### Phase 4: 搜索核心功能

- [Task 017: 测试 useSearch Hook（debounce、分页、竞态处理）](./task-017-red-use-search.md)
- [Task 018: 实现 useSearch Hook](./task-018-green-use-search.md)
- [Task 019: 测试 useDownload Hook（获取链接、缓存、状态管理）](./task-019-red-use-download.md)
- [Task 020: 实现 useDownload Hook](./task-020-green-use-download.md)
- [Task 021: 测试下载按钮组件（四态切换）](./task-021-red-download-button.md)
- [Task 022: 实现下载按钮组件](./task-022-green-download-button.md)

### Phase 5: UI 模式与交互

- [Task 023: 测试 useViewMode Hook（Hero/Docked 切换 + View Transition）](./task-023-red-use-view-mode.md)
- [Task 024: 实现 useViewMode Hook](./task-024-green-use-view-mode.md)
- [Task 025: 测试 useHotkey Hook（Cmd/Ctrl+K）](./task-025-red-use-hotkey.md)
- [Task 026: 实现 useHotkey Hook](./task-026-green-use-hotkey.md)
- [Task 027: 测试搜索输入组件（输入框 + 清空 + 快捷键徽章）](./task-027-red-search-input.md)
- [Task 028: 实现搜索输入组件](./task-028-green-search-input.md)

### Phase 6: 搜索页面集成

- [Task 029: 测试搜索页面完整流程](./task-029-red-search-page.md)
- [Task 030: 实现搜索页面](./task-030-green-search-page.md)

### Phase 7: Admin 认证与同步管理

- [Task 031: 测试 useAdminAuth Hook（localStorage + 验证 + 401 拦截）](./task-031-red-use-admin-auth.md)
- [Task 032: 实现 useAdminAuth Hook](./task-032-green-use-admin-auth.md)
- [Task 033: 测试 API Key 输入对话框组件](./task-033-red-api-key-dialog.md)
- [Task 034: 实现 API Key 输入对话框组件](./task-034-green-api-key-dialog.md)
- [Task 035: 测试 useSyncProgress Hook（轮询、启停）](./task-035-red-use-sync-progress.md)
- [Task 036: 实现 useSyncProgress Hook](./task-036-green-use-sync-progress.md)
- [Task 037: 测试同步进度展示组件](./task-037-red-sync-progress.md)
- [Task 038: 实现同步进度展示组件](./task-038-green-sync-progress.md)
- [Task 039: 测试管理页面完整流程](./task-039-red-admin-page.md)
- [Task 040: 实现管理页面](./task-040-green-admin-page.md)

### Phase 8: 路由与集成

- [Task 041: 测试路由导航（搜索页、管理页、404、search params）](./task-041-red-routing.md)
- [Task 042: 实现完整路由配置与页面集成](./task-042-green-routing.md)

### Phase 9: 可访问性与精细化

- [Task 043: 测试可访问性（ARIA 属性、键盘导航、焦点管理）](./task-043-red-accessibility.md)
- [Task 044: 实现可访问性增强](./task-044-green-accessibility.md)

### Phase 10: 构建集成与验收

- [Task 045: 构建产物集成（Vite build → Go embed.FS）](./task-045-build-integration.md)
- [Task 046: 全量验收测试与代码质量检查](./task-046-final-verification.md)

---

## Execution Handoff

**Plan complete and saved to `docs/plans/2026-02-21-react-frontend-rewrite-plan/`. Execution options:**

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
