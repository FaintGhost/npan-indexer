# Task 002: Docker Compose CI 集成

**depends-on**: (none)

## Objective

在 `docker-compose.ci.yml` 中添加 Playwright 服务容器，更新 `Makefile` 添加 `e2e-test` target，更新 GitHub Actions workflow 添加 E2E 测试 job。

## Files to Create/Modify

| File | Action |
|------|--------|
| `docker-compose.ci.yml` | 修改：添加 `playwright` service |
| `Makefile` | 修改：添加 `e2e-test` target |
| `.github/workflows/ci.yml` | 修改：添加 `e2e-test` job |

## Steps

### 1. 添加 Playwright 服务到 docker-compose.ci.yml

在 `docker-compose.ci.yml` 中添加 `playwright` service：
- `image: mcr.microsoft.com/playwright:v1.52.0-noble`
- `container_name: playwright-ci`
- `depends_on`: npan (service_healthy)
- `ipc: host` — 防止 Chromium 共享内存不足崩溃
- `init: true` — PID 1 信号处理
- `working_dir: /web`
- `volumes`: `./web:/web`
- `profiles: [e2e]` — 默认不启动，仅 e2e-test 时运行
- `environment`:
  - `BASE_URL: http://npan:1323`
  - `MEILI_HOST: http://meilisearch:7700`
  - `MEILI_API_KEY: ci-test-meili-key-5678`
  - `MEILI_INDEX: npan_items`
  - `E2E_ADMIN_API_KEY: ci-test-admin-api-key-1234`
  - `CI: "true"`
- `command: sh -c "npm install --frozen-lockfile 2>/dev/null; npx playwright test"`

**注意**：容器内使用 `npx` 而非 `bun`，因为 Playwright 官方镜像预装 Node.js 但不含 bun。

### 2. 添加 Makefile e2e-test target

添加 `e2e-test` target，流程：
1. 使用 trap 确保 cleanup（`docker compose down --volumes`）
2. `docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120`
3. 先运行 smoke test 确保服务可用
4. `docker compose -f docker-compose.ci.yml run --rm --profile e2e playwright`
5. cleanup

### 3. 更新 GitHub Actions CI workflow

在 `.github/workflows/ci.yml` 添加 `e2e-test` job：
- `needs: [smoke-test]` — 依赖冒烟测试通过
- `runs-on: ubuntu-latest`
- Steps:
  1. `actions/checkout@v4`
  2. `docker/setup-buildx-action@v3`
  3. Start app services（up --build -d --wait）
  4. Run E2E tests（docker compose run --rm --profile e2e playwright）
  5. Upload artifacts（`if: ${{ !cancelled() }}`）— `web/playwright-report/`、`web/test-results/`，retention 7 天
  6. Export logs on failure
  7. Cleanup（always）

## Verification

```bash
# 验证 docker-compose.ci.yml 语法正确
docker compose -f docker-compose.ci.yml config

# 验证 playwright service 在 e2e profile 中
docker compose -f docker-compose.ci.yml --profile e2e config | grep playwright

# 验证 Makefile target 存在
make -n e2e-test
```
