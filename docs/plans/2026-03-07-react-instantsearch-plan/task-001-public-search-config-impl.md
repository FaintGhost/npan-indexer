# Task 001: [IMPL] 公开搜索配置契约与安全边界 (GREEN)

**depends-on**: task-001-public-search-config-test.md

## Description

实现公开搜索配置契约，引入 dedicated public search config，并确保浏览器只拿到允许公开的搜索配置。该任务还需要补齐 proto 与生成链路。

## Execution Context

**Task Number**: 002 of 013
**Phase**: Foundation
**Prerequisites**: `task-001-public-search-config-test.md` 已完成且处于 Red

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

- Modify: `proto/npan/v1/api.proto`
- Modify: `internal/config/config.go`
- Modify: `internal/config/validate.go`
- Modify: `internal/httpx/connect_app_auth_search.go`
- Modify: `.env.example`
- Modify: `.env.meilisearch.example`
- Modify: `gen/go/npan/v1/*`
- Modify: `web/src/gen/**/*`

## Steps

### Step 1: Update Contract

- 在 `AppService` 下新增公开搜索配置 RPC 与响应消息。
- 明确返回字段只包含公开搜索页真正需要的配置：`host`、`indexName`、`searchApiKey`、`instantsearchEnabled`。
- 运行 Buf 生成链路，保持 Go/TS 客户端一致。

### Step 2: Implement Config Boundary

- 在服务端配置中引入 dedicated public search config 字段。
- 为新字段补齐加载、校验和日志脱敏边界。
- 禁止浏览器配置直接回落到私有 `MEILI_API_KEY`。

### Step 3: Implement AppService Handler

- 在 `AppService` 中实现公开搜索配置返回逻辑。
- 保证关闭开关或配置不完整时，返回可用于前端回退的状态，而不是暴露错误凭据。

### Step 4: Verify Green

- 运行 task-001 新增测试并确认转绿。
- 运行 Buf lint/generate 与 AppService 相关测试，确保无契约漂移。

## Verification Commands

```bash
XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf lint
XDG_CACHE_HOME=/tmp/.cache BUF_CACHE_DIR=/tmp/.cache/buf ./.bin/buf generate
GOCACHE=/tmp/go-build go test ./internal/httpx ./internal/config -run 'GetSearchConfig|PublicSearch|SearchConfig' -count=1
git diff --check
```

## Success Criteria

- 公开搜索配置 RPC 与生成产物齐备。
- dedicated public search config 与私有 Meili 配置边界清晰。
- task-001 新增测试通过。
- 不向浏览器暴露 admin/master key。
