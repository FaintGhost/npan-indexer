# Lessons

## 2026-02-24

- 用户纠正：新增能力时要确保前后端通过 `openapi.yaml` 对齐（当前项目以 `api/openapi.yaml` 为契约源）。
- 规则：
  - 在实现前先检查 `api/openapi.yaml` 是否已覆盖所需字段。
  - 若有字段变更，先改 `api/openapi.yaml`，再生成 `api/types.gen.go` 与 `web/src/api/generated/*`。
  - 实现完成前必须执行等价 `generate-check`（生成 + `git diff --exit-code`）。
- 用户纠正：新增功能后必须跑完整验证链（前后端单测、冒烟测试、E2E 测试）。
- 规则：
  - 默认收口命令至少包含：`go test ./...`、`cd web && bun vitest run`。
  - 若仓库提供 Docker 冒烟与 E2E（本项目有），必须执行并记录结果：
    - `docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120`
    - `./tests/smoke/smoke_test.sh`
    - `docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright`
    - `docker compose -f docker-compose.ci.yml --profile e2e down --volumes`
  - 若环境缺失依赖（如 `make`/Docker 权限），需明确说明并给出等价命令或申请提权后执行。
