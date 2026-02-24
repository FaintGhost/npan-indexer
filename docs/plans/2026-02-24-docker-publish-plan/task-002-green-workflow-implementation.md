# Task 002: GREEN docker publish workflow implementation

**depends-on**: task-001-red-workflow-contract.md

## Objective

实现可执行发布流水线：buildx 构建并推送到 Docker Hub 与 GHCR。

## Files

- `.github/workflows/docker-publish.yml`

## Tasks

1. 增加 `permissions.packages: write` 供 GHCR 推送。
2. 使用 `docker/login-action` 分别登录 Docker Hub 与 GHCR。
3. 使用 `docker/metadata-action` 统一生成双仓库 tags/labels。
4. 使用 `docker/build-push-action` 执行多平台构建与双仓库推送。

## Verification

- 校对 workflow 字段完整性与 action 版本稳定性
