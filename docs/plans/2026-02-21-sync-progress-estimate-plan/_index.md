# sync-full 估算进度增强计划（2026-02-21）

## Goal

- 基于 Npan 返回的目录 `item_count` 增加一个“估算总量”维度。
- 在 `sync-full` 人类可读日志中展示估算进度百分比，辅助判断同步完成度。

## Constraints

- 不改变全量同步核心遍历与断点续跑语义。
- 估算字段为可选能力：无法获取总量时回退为 `n/a`，不可阻塞同步。
- 保持现有 `sync-progress` JSON 兼容（新增字段仅追加，不破坏旧字段）。

## BDD Reference

- [BDD Specifications](./bdd-specs.md)

## Execution Plan

- [Task 001: 场景红测（估算字段与摘要渲染）](./task-001-red-estimate-progress.md)
- [Task 002: 场景绿测（接入 item_count 并输出估算进度）](./task-002-green-estimate-progress.md)
- [Task 003: 文档与验证（运行手册与全量回归）](./task-003-verify-docs-and-cli.md)

## Progress

- [Progress](./progress.md)
