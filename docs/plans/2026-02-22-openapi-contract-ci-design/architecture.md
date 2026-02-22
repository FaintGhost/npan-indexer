# Architecture

## 组件关系

```
api/openapi.yaml (Single Source of Truth)
       │
       ├── oapi-codegen ──► api/types.gen.go (Go types)
       │                        │
       │                        └── internal/httpx/handlers.go (使用生成的 types)
       │
       └── @hey-api/openapi-ts ──► web/src/api/generated/ (TS types + Zod schemas)
                                       │
                                       └── web/src/hooks/ (使用生成的 schemas)
```

## 文件结构变更

```
api/                              # 新目录
  openapi.yaml                    # OpenAPI 3.1 spec
  oapi-codegen.yaml               # Go codegen 配置
  types.gen.go                    # 生成的 Go types（DO NOT EDIT）

web/
  openapi-ts.config.ts            # TS codegen 配置
  src/api/generated/              # 生成的 TS 代码（DO NOT EDIT）
    types.gen.ts
    zod.gen.ts
  src/lib/
    schemas.ts                    # 逐步替换为 import from api/generated
    sync-schemas.ts               # 逐步替换为 import from api/generated

tests/
  smoke/
    smoke_test.sh                 # 冒烟测试脚本

docker-compose.ci.yml             # CI 专用 compose
Makefile                          # 构建工作流入口

.github/workflows/
  ci.yml                          # CI pipeline
```

## CI Pipeline 架构

```
┌──────────┐  ┌──────────────┐  ┌───────────────────┐  ┌────────────────┐
│   lint   │  │ unit-test-go │  │ unit-test-frontend │  │ generate-check │
└────┬─────┘  └──────┬───────┘  └─────────┬─────────┘  └───────┬────────┘
     │               │                    │                     │
     └───────────────┴────────────────────┴─────────────────────┘
                                  │
                          ┌───────▼───────┐
                          │  smoke-test   │
                          │ (docker compose)│
                          └───────────────┘
```

## 代码生成流程

### Go 端

1. `oapi-codegen --config api/oapi-codegen.yaml api/openapi.yaml`
2. 生成 `api/types.gen.go`，包含所有 response/request 的 Go struct
3. Handler 代码 import 并使用这些 types
4. 现有 `internal/httpx/dto.go` 中的手写 struct 逐步替换

### TS 端

1. `cd web && npx @hey-api/openapi-ts`
2. 生成 `web/src/api/generated/zod.gen.ts`（Zod schemas）和 `types.gen.ts`（TS 类型）
3. 前端代码 import 生成的 schemas 替代手写版本
4. `api-client.ts` 的 `apiGet` 使用生成的 schema 做 runtime validation

## docker-compose.ci.yml 设计

与生产 `docker-compose.yml` 的关键差异：

| 项目 | 生产 | CI |
|------|------|------|
| env 来源 | `.env` 文件 | 内联 environment |
| restart | always | 不设置 |
| volumes | 持久化 named volumes | 无 |
| healthcheck | interval 10-15s | interval 5s（更快反馈） |
| MEILI_ENV | 默认 production | development（快速启动） |
