# Task 001: [TEST] 公开搜索配置契约与安全边界 (RED)

**depends-on**: (none)

## Description

为“公开搜索配置”补充失败测试，锁定浏览器只可获得 dedicated public search config，而不能直接暴露私有 Meilisearch 凭据。该任务只负责编写和验证失败测试，不实现接口本身。

## Execution Context

**Task Number**: 001 of 013
**Phase**: Foundation
**Prerequisites**: 已阅读 `internal/httpx/connect_app_auth_search.go` 与 `internal/config/config.go` 当前 AppService/Config 结构

## BDD Scenario

```gherkin
Scenario: 页面启动时成功获取公开搜索配置
  Given 后端返回公开搜索配置 host、indexName、searchApiKey 和 instantsearchEnabled=true
  When 用户打开搜索页
  Then 前端应初始化 InstantSearch search client
  And 浏览器不应调用 AppService.AppSearch

Scenario: 浏览器只暴露 search-only key
  Given 搜索页运行在浏览器环境中
  When 页面初始化公开搜索客户端
  Then 使用的凭证必须是 search-only key
  And 页面中不得出现 Meilisearch admin 或 master key
```

**Spec Source**: `../2026-03-07-react-instantsearch-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `internal/httpx/connect_app_auth_search_test.go`
- Modify: `internal/config/validate_test.go`
- Modify: `internal/config/config_log_test.go`（如需补充日志与暴露边界测试）

## Steps

### Step 1: Verify Scenario

- 确认以上 2 个场景在设计 BDD 文档中存在且与“公开搜索配置 + search-only key”边界一致。

### Step 2: Implement Test (Red)

- 在后端测试中增加“公开搜索配置”契约断言，覆盖：
  - 存在独立的公开搜索配置返回值
  - 可返回 `host / indexName / searchApiKey / instantsearchEnabled`
  - 未启用或配置缺失时不会误暴露私有凭据
- 在配置层测试中增加 dedicated public search config 约束，避免实现时直接复用私有 `MEILI_API_KEY`。
- 使用 `httptest`、测试配置对象或其他测试替身，隔离外部 Meilisearch 和真实网络。

### Step 3: Verify Red Failure

- 运行目标 Go 测试并确认失败。
- 失败原因必须指向“缺少公开搜索配置契约/安全边界”，不能是编译环境、网络连接或依赖缺失问题。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/httpx ./internal/config -run 'GetSearchConfig|PublicSearch|SearchConfig' -count=1
```

## Success Criteria

- 新增测试可稳定运行并处于 Red。
- 失败为断言失败，且指向契约或安全边界缺失。
- 不依赖真实 Meilisearch 或外部网络。
