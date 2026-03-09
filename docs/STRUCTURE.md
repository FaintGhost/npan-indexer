# 项目结构说明

## 目录职责

- `cmd/server`：HTTP 服务入口，负责加载配置、挂载 Connect 路由并提供嵌入前端。
- `cmd/cli`：命令行工具入口，提供 token、搜索、下载、同步与进度查询。
- `internal`：核心业务实现（HTTP / service / indexer / search / config 等）。
- `data`：运行期状态目录（checkpoint / progress / dump）。
- `docs/runbooks`：当前运维与操作手册。
- `docs/reference`：外部参考资料。
- `docs/archive`：历史记录归档，不作为当前实现说明主入口。
- `docs/plans`：历史设计与实施计划，不作为当前运行时事实入口。
- `web`：前端源码、生成代码、Vitest 与 Playwright 用例；生产产物会被后端嵌入。

## 当前运行形态

- 运行时已是 Connect-only，主 RPC 路径位于 `/npan.v1.*`。
- 健康检查仍保留 HTTP：`GET /healthz`、`GET /readyz`。
- 后端通过 `web/embed.go` 中的 `//go:embed all:dist` 嵌入 `web/dist`。
- 前端包管理与脚本入口以 `web/package.json` 为准，默认使用 `bun`。

## 发布前建议

- 确认 `.env` 不在版本控制中，仅保留 `.env.example`。
- 确认 `NPA_ADMIN_API_KEY` 已配置且长度不少于 16 字符。
- 若计划让管理接口回退使用服务端凭据，显式确认 `NPA_ALLOW_CONFIG_AUTH_FALLBACK=true`，并同时提供 `NPA_TOKEN` 或完整 OAuth 三元组。
- 若启用浏览器公开搜索，确认 `MEILI_PUBLIC_*` 已配置完整，且 `MEILI_PUBLIC_SEARCH_API_KEY` 为 dedicated search-only key。
- 运行验证：
  - `make test`
  - `make test-frontend`
  - `cd web && bun run typecheck`
  - `make smoke-test`
  - `make e2e-test`
- 若改了 protobuf 契约，额外执行：
  - `buf lint`
  - `buf generate`
- 若部署在反向代理后，再根据实际网络拓扑评估是否调整 `IPExtractor`。
