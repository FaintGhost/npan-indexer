# Task 008: RED frontend estimate warning display tests

**depends-on**: task-005-green-backend-folder-info-estimate-and-warnings.md

## Scenario Reference

- Feature: Completion Warning For Estimate Mismatch
- Scenario: 实际写入显著小于官方估计时生成警告
- Scenario: 差异在阈值内不告警

## Objective

先补 UI 红测，确保差异告警显示行为可回归验证。

## Files

- `web/src/components/sync-progress-display.test.tsx`

## Tasks

1. 新增测试：verification.warnings 包含估计差异时展示“验证警告”区域。
2. 新增测试：无估计差异 warning 时不展示该告警内容。
3. 新增测试：root 详情中显示 `estimatedTotalDocs` 与实际统计（若有）。

## Verification

- `cd web && bun vitest run src/components/sync-progress-display.test.tsx`
- 预期：新增测试在实现前失败（RED）
