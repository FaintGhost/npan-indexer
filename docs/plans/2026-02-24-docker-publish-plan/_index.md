# Docker Publish Plan

## Goal

为当前项目增加 GitHub Actions 发布流水线：构建容器镜像并推送到 Docker Hub 与 GHCR。

## Constraints

- 使用官方成熟 Action（`docker/*-action`）实现，避免自写脚本拼装登录与打标。
- 推送动作仅在 `main` / tag / 手动触发执行，避免 PR 污染镜像仓库。
- 标签策略需包含 `latest`（默认分支）、分支/标签引用与 `sha`。

## Execution Plan

- [Task 001: RED workflow contract and trigger definition](./task-001-red-workflow-contract.md)
- [Task 002: GREEN docker publish workflow implementation](./task-002-green-workflow-implementation.md)
- [Task 003: GREEN documentation for secrets and usage](./task-003-green-docs.md)
- [Task 004: Verification and risk check](./task-004-verification.md)
