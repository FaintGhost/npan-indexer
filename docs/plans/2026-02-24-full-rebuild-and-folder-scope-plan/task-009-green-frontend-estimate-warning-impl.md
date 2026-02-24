# Task 009: GREEN frontend estimate warning display

**depends-on**: task-008-red-frontend-estimate-warning-tests.md

## Scenario Reference

- Feature: Completion Warning For Estimate Mismatch
- Scenario: 实际写入显著小于官方估计时生成警告
- Scenario: 差异在阈值内不告警

## Objective

在同步进度 UI 中展示 root 级统计预估与完成差异提示，帮助快速识别异常目录。

## Files

- `web/src/components/sync-progress-display.tsx`

## Tasks

1. 在 root 详情处显示 `estimatedTotalDocs`（仅存在时显示）。
2. 计算并展示 root 实际文档数（`filesIndexed + foldersVisited`）用于人工对照。
3. 复用 `verification.warnings` 展示完成后的差异警告内容，避免引入新字段。

## Verification

- `cd web && bun vitest run src/components/sync-progress-display.test.tsx`
