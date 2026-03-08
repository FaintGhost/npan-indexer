# Task 002: [TEST] 搜索页配置引导与回退 (RED)

**depends-on**: (none)

## Description

为搜索页增加失败测试，覆盖：启动时加载公开搜索配置、开启时创建直连搜索客户端、关闭或配置缺失时回退到旧 `AppSearch` 链路。该任务不实现生产代码。

## Execution Context

**Task Number**: 003 of 013
**Phase**: Foundation
**Prerequisites**: 理解当前 `web/src/routes/index.lazy.tsx` 的 `AppSearch` 状态机，以及 `web/src/components/search-page.test.tsx` 的 MSW 测试结构

## BDD Scenario

```gherkin
Scenario: 页面启动时成功获取公开搜索配置
  Given 后端返回公开搜索配置 host、indexName、searchApiKey 和 instantsearchEnabled=true
  When 用户打开搜索页
  Then 前端应初始化 InstantSearch search client
  And 浏览器不应调用 AppService.AppSearch

Scenario: 公开搜索配置不可用时回退旧链路
  Given 后端返回 instantsearchEnabled=false 或公开搜索配置缺失
  When 用户打开搜索页
  Then 前端应回退到现有 AppSearch 链路
  And 页面仍可完成搜索与下载

Scenario: 直连链路异常时可切回 AppSearch
  Given 预发或线上发现直连 Meilisearch 存在不可接受的问题
  When 运维关闭 instantsearchEnabled 开关
  Then 前端应切回 AppSearch 搜索链路
  And 下载与页面其他功能保持可用
```

**Spec Source**: `../2026-03-07-react-instantsearch-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/src/components/search-page.test.tsx`
- Modify: `web/src/tests/mocks/handlers.ts`
- Create: `web/src/lib/search-config.test.ts`（如需要把配置加载与回退逻辑拆成独立可测单元）

## Steps

### Step 1: Verify Scenario

- 确认 3 个场景都在 BDD 设计中存在，且语义聚焦“配置引导 + fallback”。

### Step 2: Implement Test (Red)

- 使用 MSW 作为网络测试替身，模拟：
  - `GetSearchConfig` 开启状态
  - `GetSearchConfig` 关闭状态
  - 配置缺失或异常状态
  - 旧 `AppSearch` 搜索请求
- 使用模块 mock 或等价测试替身观察“公开搜索 client 工厂是否被调用”。
- 编写组件测试，验证开启与关闭两条链路的引导分支。

### Step 3: Verify Red Failure

- 运行目标前端测试并确认新增用例失败。
- 失败原因应为“当前页面没有做配置引导或 fallback”，而不是测试环境未初始化。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/lib/search-config.test.ts
```

## Success Criteria

- 新增用例稳定失败（Red）。
- 使用 MSW / module mock 隔离网络与 client 工厂。
- 失败信息明确指向“配置引导/回退未实现”。
