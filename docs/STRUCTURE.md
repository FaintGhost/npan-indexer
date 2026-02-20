# 项目结构说明

## 目录职责

- `cmd/server`：HTTP API 服务入口。
- `cmd/cli`：命令行工具入口。
- `internal`：核心业务实现（按领域拆分）。
- `data`：运行期状态目录（checkpoint/progress/dump）。
- `docs/runbooks`：运维与操作手册。
- `docs/reference`：外部参考资料。
- `docs/archive`：历史记录归档（含旧技术栈 legacy 文档）。

## 发布前建议

- 确认 `.env` 不在版本控制中，仅保留 `.env.example`。
- 运行验证：
  - `go test ./...`
  - `go test -race ./...`
  - `go build ./...`
- 若开放公网接口，配置 `NPA_ADMIN_API_KEY`。
- 默认保持 `NPA_ALLOW_CONFIG_AUTH_FALLBACK=false`。
