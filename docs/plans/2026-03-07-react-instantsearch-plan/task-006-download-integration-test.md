# Task 006: [TEST] 直连搜索结果下载集成 (RED)

**depends-on**: task-004-instantsearch-results-impl.md

## Description

为直连搜索结果下的下载链路补充失败测试，锁定“搜索直连 Meilisearch，但下载仍必须经过 `AppService.AppDownloadURL`”这一安全边界与 UI 集成行为。该任务不实现生产代码。

## Execution Context

**Task Number**: 011 of 013
**Phase**: Integration
**Prerequisites**: `task-004-instantsearch-results-impl.md` 已完成，且搜索结果已切换为 InstantSearch hits

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

- Modify: `web/src/hooks/use-download.test.ts`
- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/tests/mocks/handlers.ts`
- Modify: `web/e2e/tests/search.spec.ts`
- Modify: `web/e2e/pages/search-page.ts`

## Steps

### Step 1: Verify Scenario

- 明确该任务只验证下载边界，不验证搜索结果本身的召回与排序正确性。
- 确认新结果模型只提供下载所需标识，不应承担下载 URL 分发职责。

### Step 2: Implement Test (Red)

- 在 hook 与组件测试中增加“点击 InstantSearch 结果下载按钮后仍调用 `AppDownloadURL`”的失败断言。
- 在 E2E 中观察实际网络，确保下载请求仍命中 Connect RPC，而不是从 hit 数据直接取下载地址。
- 增加负向断言，锁定公开 hit 数据结构不应要求内嵌下载 URL。

### Step 3: Verify Red Failure

- 运行目标测试并确认失败。
- 失败原因应指向 InstantSearch 结果与下载按钮尚未正确桥接，而不是 `useDownload` 本身损坏。

## Verification Commands

```bash
cd web && bun vitest run src/hooks/use-download.test.ts src/components/search-page.test.tsx
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright bunx playwright test web/e2e/tests/search.spec.ts --grep "下载"
```

## Success Criteria

- 新增下载集成用例稳定失败（Red）。
- 失败明确指向 InstantSearch 结果尚未正确复用 `AppDownloadURL` 链路。
- 测试覆盖单测与 E2E 两层。
