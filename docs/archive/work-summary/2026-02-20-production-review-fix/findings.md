# Findings（2026-02-20）

## 审查结论摘要
1. `internal/httpx/handlers.go` 存在服务端凭据回退，导致未授权调用可借服务端身份请求上游。
2. `internal/service/sync_manager.go` 在多 root 场景下每 root 新建 limiter，导致总限流放大。
3. `internal/storage/json_store.go` 直接覆盖写 JSON，存在崩溃时文件损坏风险。
4. `internal/service/sync_manager.go` 首个 root 失败后未快速取消其他 goroutine。
5. 缺少针对关键路径的自动化测试。

## 修复结果
- 安全
  - 新增 `NPA_ADMIN_API_KEY`，可对 `/api/v1/*` 启用 API Key 保护。
  - 新增 `NPA_ALLOW_CONFIG_AUTH_FALLBACK`（默认 `false`），默认不再对 HTTP 请求回退服务端凭据。
- 稳定性
  - 同步任务改为共享全局 limiter，避免多 root 限流放大。
  - 首个 root 失败时触发 `cancel`，其余 goroutine 尽快停止。
  - JSON 持久化改为原子写（临时文件 + fsync + rename）。
- 回归保障
  - 新增 `internal/httpx/handlers_test.go`。
  - 新增 `internal/storage/json_store_test.go`。
