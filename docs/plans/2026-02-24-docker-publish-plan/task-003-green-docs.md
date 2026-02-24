# Task 003: GREEN documentation for secrets and usage

**depends-on**: task-002-green-workflow-implementation.md

## Objective

补充发布流水线的仓库配置说明，确保可落地。

## Files

- `README.md`

## Tasks

1. 增加所需 Secrets 列表：
  - `DOCKERHUB_USERNAME`
  - `DOCKERHUB_TOKEN`
2. 说明 GHCR 使用 `GITHUB_TOKEN` 推送（需 packages:write）。
3. 说明触发条件与标签规则。

## Verification

- 文档命令和字段名称与 workflow 一致
