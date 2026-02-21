# Task 022: 实现下载按钮组件

**depends-on**: task-021

## Description

实现 DownloadButton 组件，使 Task 021 测试通过。

## Execution Context

**Task Number**: 022 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 021 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 2 - 下载按钮所有状态场景

## Files to Modify/Create

- Create: `cli/src/components/download-button.tsx`

## Steps

### Step 1: Implement DownloadButton

- Props: status ("idle" | "loading" | "success" | "error"), onDownload, fileName
- 根据 status 渲染不同的图标和文字
- loading 和 success 时 disabled
- 使用 aria-label 包含文件名
- 样式匹配现有 HTML 中 BtnStates

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/components/download-button.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 021 所有测试通过
