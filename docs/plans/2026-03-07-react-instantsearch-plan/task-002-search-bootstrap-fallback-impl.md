# Task 002: [IMPL] 搜索页配置引导与回退 (GREEN)

**depends-on**: task-001-public-search-config-impl.md, task-002-search-bootstrap-fallback-test.md

## Description

为搜索页接入运行时公开搜索配置，按开关决定使用 InstantSearch 直连链路还是旧 `AppSearch` 链路，并完成依赖引入与前端初始化。

## Execution Context

**Task Number**: 004 of 013
**Phase**: Foundation
**Prerequisites**: `task-001-public-search-config-impl.md` 与 `task-002-search-bootstrap-fallback-test.md` 已完成

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

- Modify: `web/package.json`
- Create: `web/src/lib/search-config.ts`
- Create: `web/src/lib/meili-search-client.ts`
- Modify: `web/src/routes/index.lazy.tsx`
- Modify: `web/src/tests/mocks/handlers.ts`（若共享测试数据结构需要同步）

## Steps

### Step 1: Add Official Dependencies

- 引入官方 `react-instantsearch`、`@meilisearch/instant-meilisearch` 与所需样式依赖。
- 确保依赖安装方式和项目现有 Bun 工作流一致。

### Step 2: Implement Runtime Bootstrap

- 新建公开搜索配置读取模块，负责从 `GetSearchConfig` 获取并校验运行时配置。
- 新建直连 Meili search client 工厂，并保证实例创建可被测试替身观测。
- 在搜索页入口根据 `instantsearchEnabled` 选择直连或 legacy `AppSearch` 路径。

### Step 3: Preserve Fallback Path

- 保留旧 `AppSearch` 搜索路径与下载能力。
- 当公开搜索配置关闭、缺失或不可用时，页面自动走 legacy 路径，不破坏当前可用性。

### Step 4: Verify Green

- 运行 task-002 新增测试并确认通过。
- 运行搜索页与下载相关基础测试，确认回退链路无回归。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/lib/search-config.test.ts src/hooks/use-download.test.ts
cd web && bun vitest run
```

## Success Criteria

- 搜索页可按运行时配置选择新旧两条链路。
- 公开搜索 client 只在 enabled 分支初始化。
- 关闭直连开关时页面仍可用。
- task-002 新增测试通过。
