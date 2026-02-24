# Task 004: Verification and risk check

**depends-on**: task-003-green-docs.md

## Objective

完成工作流配置后的自检，并记录无法在本地完全验证的部分。

## Files

- `.github/workflows/docker-publish.yml`
- `README.md`

## Tasks

1. 检查 YAML 语法和关键字段（触发、权限、登录、metadata、build-push）。
2. 检查镜像命名是否与仓库 owner/repo 推导一致。
3. 记录运行前置条件（GitHub Secrets / Docker Hub 仓库权限）。

## Verification

- `git diff -- .github/workflows/docker-publish.yml README.md`
- 人工审查 action 配置
