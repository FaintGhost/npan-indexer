# Task 001: [IMPL] URL 筛选状态初始化与兜底 (GREEN)

**depends-on**: task-001-url-filter-state-test.md

## Description

实现搜索页对 URL `ext` 参数的读取、合法化和默认回退能力，使页面首次渲染即可恢复正确筛选状态。

## Execution Context

**Task Number**: 002 of 012  
**Phase**: Core Features  
**Prerequisites**: `task-001-url-filter-state-test.md` 已完成并处于 Red 失败状态

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

- Modify: `web/src/routes/index.lazy.tsx`
- (Optional) Modify: `web/src/tests/test-providers.tsx`

## Steps

### Step 1: Implement Logic (Green)
- 在搜索页引入 `ext` 的 URL 状态读取。
- 对读取值执行白名单校验并归一化到合法枚举。
- 缺省和非法值统一回退为 `all`。

### Step 2: Verify Green
- 运行 task-001 对应测试并确认全部通过。

### Step 3: Regression Check
- 运行搜索页现有基础测试，确认初始态/错误态等行为未回归。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- task-001 新增用例从 Red 变为 Green。
- URL 初始化行为稳定可复现。
- 不引入后端请求改动。
