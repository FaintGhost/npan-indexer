# Task 001: [TEST] URL 筛选状态初始化与兜底 (RED)

**depends-on**: (none)

## Description

为搜索页补充 URL 筛选状态相关失败测试，覆盖默认值、合法值恢复、非法值兜底，确保后续实现有明确行为边界。

## Execution Context

**Task Number**: 001 of 012  
**Phase**: Core Features  
**Prerequisites**: 现有 `SearchPage` 测试可运行，MSW 可拦截 `AppSearch` 请求

## BDD Scenario

```gherkin
Scenario: 默认进入页面时使用全部筛选
  Given 用户访问搜索页且 URL 不包含 ext 参数
  When 页面完成初始化
  Then 筛选值应为 all
  And 结果列表显示当前已加载的全部结果

Scenario: URL 中 ext 合法值可恢复筛选状态
  Given 用户访问 /?q=mx40&ext=doc
  When 搜索结果加载完成
  Then 文档筛选应为选中状态
  And 仅展示文档类扩展名结果

Scenario: URL 中 ext 非法值会回退到 all
  Given 用户访问 /?q=mx40&ext=unknown
  When 页面完成初始化
  Then 筛选值应回退为 all
  And 页面不应报错
```

**Spec Source**: `../2026-02-27-search-results-filter-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/components/search-page.test.tsx`
- (Optional) Modify: `web/src/tests/test-providers.tsx`（若需注入路由 search 参数上下文）

## Steps

### Step 1: Verify Scenario
- 确认以上 3 个场景在设计 BDD 文档中存在且语义一致。

### Step 2: Implement Test (Red)
- 在 `search-page.test.tsx` 新增 URL `ext` 相关用例：缺省、合法、非法。
- 用 MSW 作为网络测试替身，隔离外部服务。
- 保证失败原因来自断言（行为不匹配），不是模块缺失或环境错误。

### Step 3: Verify Red Failure
- 运行目标测试并确认新增用例失败。
- 失败信息应明确指向“筛选状态未按 URL 初始化/兜底”。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- 新增测试可稳定运行并失败（Red）。
- 失败为断言失败，非 ImportError/配置错误。
- 现有无关测试不受影响。
