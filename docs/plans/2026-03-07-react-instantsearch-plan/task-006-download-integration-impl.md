# Task 006: [IMPL] 直连搜索结果下载集成 (GREEN)

**depends-on**: task-006-download-integration-test.md

## Description

将 InstantSearch 结果项与现有 `useDownload` 链路正确接通，确保公开搜索迁移后下载仍保留后端受控与状态反馈行为。

## Execution Context

**Task Number**: 012 of 013
**Phase**: Integration
**Prerequisites**: `task-006-download-integration-test.md` 已完成且处于 Red

## BDD Scenario

```gherkin
Scenario: 搜索结果下载仍通过 AppDownloadURL
  Given 用户在搜索结果中点击下载按钮
  When 前端发起下载动作
  Then 前端应调用 AppService.AppDownloadURL
  And 不应尝试从 Meilisearch 响应中直接获取下载地址
```

**Spec Source**: `../2026-03-07-react-instantsearch-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/routes/index.lazy.tsx`
- Modify: `web/src/components/search-results.tsx`
- Modify: `web/src/components/file-card.tsx`
- Modify: `web/src/hooks/use-download.ts`
- Modify: `web/src/hooks/use-download.test.ts`
- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/e2e/tests/search.spec.ts`
- Modify: `web/e2e/pages/search-page.ts`

## Steps

### Step 1: Preserve Download Boundary

- 保持下载入口继续调用 `AppService.AppDownloadURL`。
- 明确 Meilisearch hit 只提供下载所需标识，不承担下载地址分发职责。

### Step 2: Bridge Hits to Download Hook

- 让 InstantSearch 结果卡片继续复用 `useDownload` 状态机与按钮反馈。
- 确保 hit 标识到下载参数的映射稳定，不因新结构导致按钮失效。

### Step 3: Keep Existing UX Guarantees

- 保留下载中的 loading、成功、失败与重复点击保护等现有体验。
- 确保多个结果并发下载仍可独立反馈状态。

### Step 4: Verify Green

- 运行 task-006 新增测试并确认通过。
- 回归下载相关现有测试与 E2E，确认搜索链路迁移未影响下载体验。

## Verification Commands

```bash
cd web && bun vitest run src/hooks/use-download.test.ts src/components/search-page.test.tsx src/components/file-card.test.tsx
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright bunx playwright test web/e2e/tests/search.spec.ts --grep "下载"
cd web && bun vitest run
```

## Success Criteria

- InstantSearch 结果下载仍通过 `AppDownloadURL`。
- 不依赖 Meilisearch hit 中的下载地址。
- 现有下载按钮状态机无回归。
- task-006 新增测试通过。
