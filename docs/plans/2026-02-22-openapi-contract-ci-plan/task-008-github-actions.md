# Task 008: Create GitHub Actions Workflow

**depends-on**: Task 006, Task 007

**ref**: BDD Scenario "Spec 变更后未重新生成代码时 CI 失败", "服务栈正常启动"

## Description

创建 GitHub Actions CI workflow，在 push/PR 时自动运行代码生成检查、单元测试和冒烟测试。

## What to do

1. 创建 `.github/workflows/ci.yml`，包含以下 jobs：

   **并行 jobs（快速反馈）**：
   - **unit-test-go**: checkout + setup-go (cache) + `go test ./... -short -count=1 -race`
   - **unit-test-frontend**: checkout + setup bun + `cd web && bun install && bun run test`
   - **generate-check**: checkout + setup-go + setup node + install tools + `make generate-check`

   **串行 job（依赖上述全部通过）**：
   - **smoke-test**: checkout + setup-buildx + `docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120` + `./tests/smoke/smoke_test.sh` + 失败时导出日志 + always cleanup

2. 配置 `concurrency` 使同一 branch 的 CI 不重复执行

3. 配置 `paths-ignore` 排除 `*.md` 和 `docs/**` 文件的变更

4. Docker 层缓存使用 Buildx + `type=gha,mode=max`

## Files to create

- `.github/workflows/ci.yml`

## Verification

- Push 到 branch 后 GitHub Actions 能触发
- generate-check job 能检测到 spec 与生成代码不一致
- smoke-test job 能成功启动服务栈并通过冒烟测试
- 失败时日志能正常导出
