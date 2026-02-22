# Task 001: Write OpenAPI Spec

**depends-on**: none

**ref**: BDD Scenario "从 spec 生成的 Go types 与现有 handler 响应结构匹配", "从 spec 生成的 Zod schemas 能解析后端 JSON 响应"

## Description

创建 `api/openapi.yaml`，定义所有公开 API 端点的请求/响应结构。这是整个契约体系的单一信源。

## What to do

1. 创建 `api/` 目录
2. 编写 `api/openapi.yaml`（OpenAPI 3.1），包含：

### Paths

| 路径 | Method | operationId | 认证 |
|------|--------|-------------|------|
| `/healthz` | GET | healthCheck | 无 |
| `/readyz` | GET | readinessCheck | 无 |
| `/api/v1/app/search` | GET | appSearch | EmbeddedAuth |
| `/api/v1/app/download-url` | GET | appDownloadUrl | EmbeddedAuth |
| `/api/v1/token` | POST | createToken | API Key |
| `/api/v1/search/remote` | GET | remoteSearch | API Key |
| `/api/v1/search/local` | GET | localSearch | API Key |
| `/api/v1/download-url` | GET | downloadUrl | API Key |
| `/api/v1/admin/sync` | GET | getSyncProgress | API Key |
| `/api/v1/admin/sync` | POST | startSync | API Key |
| `/api/v1/admin/sync` | DELETE | cancelSync | API Key |

### Components/Schemas

从现有代码派生，确保 JSON 字段名与现有代码一致：

- **ErrorResponse**: 参考 `internal/httpx/errors.go` — `code`, `message`, `request_id`
- **IndexDocument**: 参考 `internal/models/models.go` `IndexDocument` struct — 使用 snake_case（`doc_id`, `source_id`...）
- **SearchResponse**: 参考 `internal/search/query_service.go` `QueryResult` — `items`, `total`
- **DownloadURLResponse**: `file_id`, `download_url`
- **CrawlStats**: 参考 `models.CrawlStats` — camelCase（`foldersVisited`, `filesIndexed`...）
- **RootProgress**: 参考 `models.RootSyncProgress` — camelCase（`rootFolderId`, `status`...）
- **IncrementalSyncStats**: 参考 `models.IncrementalSyncStats`
- **SyncVerification**: 参考 `models.SyncVerification`
- **SyncProgressState**: 参考 `models.SyncProgressState` — **status 字段必须为 enum**: `[idle, running, done, error, cancelled, interrupted]`
- **SyncStartRequest**: 参考 `internal/httpx/handlers.go` `syncStartPayload`
- **MessageResponse**: `{ message: string }` — 用于 POST/DELETE 成功响应

### 关键注意事项

- `SyncProgressState` 的 JSON 字段名是 camelCase（`startedAt`, `updatedAt`），与 Go struct tags 一致
- `IndexDocument` 的 JSON 字段名是 snake_case（`doc_id`, `source_id`），与 Go struct tags 一致
- 可选字段用 `required` 数组控制，对应 Go 的 `omitempty` tag
- nullable 字段（如 `activeRoot`）使用 OpenAPI 3.1 的 `type: ["integer", "null"]`

## Files to create

- `api/openapi.yaml`

## Verification

- `api/openapi.yaml` 能通过 OpenAPI lint（如 `redocly lint` 或在线验证器）
- 所有端点的路径、方法与 `internal/httpx/server.go` 中的路由注册一致
- 所有 schema 的字段名和类型与现有 Go struct 的 JSON tags 一致
