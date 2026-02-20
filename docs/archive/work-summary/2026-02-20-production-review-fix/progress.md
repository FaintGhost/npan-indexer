# 进度日志（2026-02-20）

## 已完成
- 建立会话文档：`task_plan.md`、`findings.md`、`progress.md`。
- 完成安全修复：
  - `internal/config/config.go` 新增 `NPA_ADMIN_API_KEY` 与 `NPA_ALLOW_CONFIG_AUTH_FALLBACK` 配置项。
  - `internal/httpx/handlers.go` 新增 API 访问保护，认证回退逻辑改为可控开关（默认关闭）。
- 完成稳定性修复：
  - `internal/service/sync_manager.go` 使用全局共享 limiter。
  - `internal/service/sync_manager.go` 在首个错误发生后取消其余 goroutine。
  - `internal/storage/json_store.go` 改为原子写文件。
- 完成测试补充：
  - `internal/httpx/handlers_test.go`
  - `internal/storage/json_store_test.go`

## 验证记录
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./...` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test -race ./...` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go build ./...` 通过。
