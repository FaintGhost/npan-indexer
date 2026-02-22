# OpenAPI Contract + CI Smoke Test Design

## Context

npan 项目的前端（React + Zod）和后端（Go + Echo v5）测试完全隔离，没有跨边界的契约验证。过去多次出现：

- 后端新增状态值（`interrupted`），前端 Zod schema 缺失导致 parse 失败
- 后端返回 404，前端未处理导致页面阻塞
- 后端异步启动同步，前端假设同步调用导致 UI 不更新

**根本原因**：API 契约只存在于两端各自的代码中（Go struct tags + Zod schemas），没有单一信源。

## Requirements

1. **单一信源**：一份 OpenAPI spec 定义所有 API 的请求/响应结构
2. **代码生成**：从 spec 生成 Go types 和 TS Zod schemas，替代手写
3. **CI 门禁**：生成代码与 spec 不同步时阻止合并
4. **冒烟测试**：CI 中用 Docker Compose 启动完整服务栈，验证核心端点可达
5. **低侵入**：不大幅重构现有 handler 代码

## Rationale

### 为什么选 Spec-first

- Go struct 的 json tag 和前端 Zod schema 都从 spec 派生，任何字段变更必须先改 spec
- `oapi-codegen` 只生成 types（不生成 Echo v5 server 接口，因为 oapi-codegen 尚未支持 Echo v5）
- `@hey-api/openapi-ts` 的 Zod v3 插件直接生成可用的 Zod schemas

### 为什么要 CI 冒烟测试

- 契约测试验证的是数据结构一致性，但无法覆盖时序问题（如异步启动同步）
- 冒烟测试验证"完整服务栈能启动且核心端点返回正确状态码"

## Detailed Design

### 1. OpenAPI Spec

单文件 `api/openapi.yaml`，覆盖所有公开端点：

| 分组 | 端点 | Method |
|------|------|--------|
| Public | `/healthz`, `/readyz` | GET |
| App | `/api/v1/app/search`, `/api/v1/app/download-url` | GET |
| Admin | `/api/v1/admin/sync` | GET/POST/DELETE |
| API | `/api/v1/token`, `/api/v1/search/remote`, `/api/v1/search/local`, `/api/v1/download-url` | GET/POST |

所有 schemas 定义在 `components/schemas` 中，包括：
- `ErrorResponse`（code, message, request_id）
- `SearchResponse`（items, total）
- `IndexDocument`（doc_id, source_id, type, name...）
- `SyncProgressState`（status enum, roots, aggregateStats...）
- `CrawlStats`, `RootProgress`, `IncrementalSyncStats`, `SyncVerification`
- `DownloadURLResponse`（file_id, download_url）

**关键约束**：`SyncProgressState.status` 定义为 enum `[idle, running, done, error, cancelled, interrupted]`，后端新增状态值时必须先改 spec。

### 2. Go 代码生成

工具：`oapi-codegen v2.5+`，仅生成 types。

```yaml
# api/oapi-codegen.yaml
package: api
output: api/types.gen.go
generate:
  models: true
  strict-server: false
  echo-server: false
  client: false
  embedded-spec: false
```

生成的 types 用于：
- handler 中替代手写 response struct（如 `dto.go` 中的 `SyncProgressResponse`）
- handler 测试中做类型断言

### 3. TypeScript 代码生成

工具：`@hey-api/openapi-ts`，生成 Zod schemas + TS types。

```typescript
// web/openapi-ts.config.ts
export default defineConfig({
  input: "../api/openapi.yaml",
  output: "src/api/generated",
  plugins: [
    "@hey-api/typescript",
    { name: "zod", compatibilityVersion: 3 },
  ],
})
```

生成后替换手写的 `web/src/lib/schemas.ts` 和 `web/src/lib/sync-schemas.ts`。

### 4. 构建工作流

`Makefile` 提供以下目标：

- `make generate` — 一键生成 Go + TS 代码
- `make generate-check` — CI 用，检查生成代码是否与 spec 同步
- `make test` — 运行 Go 单元测试
- `make test-frontend` — 运行前端测试
- `make smoke-test` — Docker Compose 冒烟测试

### 5. CI Pipeline（GitHub Actions）

```
push/PR to main
  ├─ lint (parallel)
  ├─ unit-test-go (parallel)
  ├─ unit-test-frontend (parallel)
  ├─ generate-check (parallel)
  └─ smoke-test (depends on all above)
```

冒烟测试使用独立的 `docker-compose.ci.yml`（内联环境变量，不依赖 .env 文件），通过 Shell + curl 脚本验证核心端点。

### 6. 现有代码适配

- `GetFullSyncProgress` handler 当前直接返回 `models.SyncProgressState`，应改为使用生成的 response type（或通过 DTO 映射）
- 前端 `api-client.ts` 的 `apiGet` 接受 Zod schema 参数，替换为生成的 schema 即可
- 手写的 `schemas.ts` 和 `sync-schemas.ts` 标记为 deprecated，逐步替换为生成代码

## Design Documents

- [BDD Specifications](./bdd-specs.md) - 行为场景和测试策略
- [Architecture](./architecture.md) - 系统架构和组件细节
- [Best Practices](./best-practices.md) - 安全、性能和代码质量指南
