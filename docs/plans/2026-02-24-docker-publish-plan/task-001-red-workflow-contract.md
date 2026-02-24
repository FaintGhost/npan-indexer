# Task 001: RED workflow contract and trigger definition

**depends-on**: none

## Objective

明确工作流触发条件、镜像命名与标签策略，避免后续实现返工。

## Files

- `.github/workflows/docker-publish.yml`

## Tasks

1. 定义触发条件：`push` 到 `main`、版本 tag、`workflow_dispatch`。
2. 定义镜像目标：
  - Docker Hub：`<DOCKERHUB_USERNAME>/<repo>`
  - GHCR：`ghcr.io/<owner>/<repo>`
3. 定义标签策略：`latest`、`ref`、`sha`。

## Verification

- 自查 YAML 结构与触发范围（不实现推送逻辑，仅明确契约）
