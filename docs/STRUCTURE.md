# 项目结构说明

## 目录职责

- `cmd/server`：HTTP API 服务入口
- `cmd/cli`：命令行工具入口
- `internal`：核心业务实现
- `data`：运行期数据目录（默认不提交内容）
- `docs/archive`：历史方案与阶段记录归档
- `docs/reference`：参考资料
- `docs/runbooks`：运行手册

## 发布建议

- 提交前确认 `.env` 未纳入版本控制。
- 运行 `go test ./...` 与 `go build ./...`。
- 若公开仓库，确认所有密钥通过环境变量注入，不写死在配置文件中。
