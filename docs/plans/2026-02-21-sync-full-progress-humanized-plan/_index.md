# sync-full 进度输出可读性优化计划（2026-02-21）

## Goal

- 将 `sync-full` 的轮询进度输出从原始 JSON 切换为更适合人类阅读的一行摘要格式。
- 保留机器可消费能力，提供可切换的 JSON 进度输出模式。

## Constraints

- 不改变同步核心流程与状态机行为，仅优化 CLI 展示层。
- 保持最终结果输出与错误语义兼容。
- 通过自动化测试覆盖格式化与模式切换逻辑。

## BDD Reference

- [BDD Specifications](./bdd-specs.md)

## Execution Plan

- [Task 001: 场景红测（进度格式化与输出模式）](./task-001-red-progress-format.md)
- [Task 002: 场景绿测（实现人类可读进度输出）](./task-002-green-progress-format.md)
- [Task 003: 文档与验证（运行手册与命令校验）](./task-003-verify-docs-and-cli.md)

## Progress

- [Progress](./progress.md)
