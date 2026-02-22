# Task 007: Create CI Compose and Smoke Test

**depends-on**: none

**ref**: BDD Scenarios "服务栈正常启动", "未认证请求被拒绝", "认证后管理端点可用", "搜索端点返回正确结构"

## Description

创建 CI 专用的 Docker Compose 文件和冒烟测试脚本。

## What to do

### docker-compose.ci.yml

创建 `docker-compose.ci.yml`，与生产版本的关键差异：
- 环境变量内联（不依赖 .env 文件）
- 不设置 `restart`
- 无持久化 volumes
- healthcheck interval 5s（加速启动检测）
- Meilisearch 使用 `MEILI_ENV=development`
- `NPA_ADMIN_API_KEY` 设为固定测试值（如 `ci-test-admin-api-key-1234`）
- `MEILI_MASTER_KEY` 和 `MEILI_API_KEY` 使用相同的固定测试值
- `NPA_ALLOW_CONFIG_AUTH_FALLBACK: "true"` 以便 app 端点不需要真实凭据即可测试

### smoke_test.sh

创建 `tests/smoke/smoke_test.sh`，验证以下场景：

1. **健康检查**：`GET /healthz` 返回 200，JSON 包含 `{"status": "ok"}`
2. **就绪检查**：`GET /readyz` 返回 200
3. **未认证拒绝**：`GET /api/v1/admin/sync` 不带 key 返回 401
4. **认证后可用**：`GET /api/v1/admin/sync` 带正确 key 返回 200 或 404
5. **搜索端点**：`GET /api/v1/app/search?q=test` 返回 200，JSON 包含 `items` 和 `total`
6. **Metrics 端点**：`GET http://localhost:9091/metrics` 返回 200

脚本使用 curl + jq 做断言，输出 PASS/FAIL 汇总，失败时 exit 1。

## Files to create

- `docker-compose.ci.yml`
- `tests/smoke/smoke_test.sh`（需要 `chmod +x`）

## Verification

- 在本地执行 `docker compose -f docker-compose.ci.yml up -d --wait` 服务能正常启动
- `./tests/smoke/smoke_test.sh` 全部 PASS
- `docker compose -f docker-compose.ci.yml down --volumes` 正常清理
