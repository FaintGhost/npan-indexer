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
  - `make test`（或 `go test ./...`）
  - `go test -race ./...`
  - `make smoke-test`（Docker 34 项冒烟测试）
  - `make e2e-test`（Playwright 32 项 E2E 测试）
- 若开放公网接口，配置 `NPA_ADMIN_API_KEY`（>= 16 字符，空值启动时 panic）。
- 默认保持 `NPA_ALLOW_CONFIG_AUTH_FALLBACK=false`。
- 若部署在反向代理后，将 `IPExtractor` 改为 `ExtractIPFromXFFHeader()`。
