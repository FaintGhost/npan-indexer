# Task 021: 测试下载按钮组件（四态切换）

**depends-on**: task-004

## Description

为 DownloadButton 组件创建失败测试用例。测试四种按钮状态的渲染。

## Execution Context

**Task Number**: 021 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 2 - 下载按钮加载状态; 下载链接获取成功后打开新标签页; 下载链接获取失败显示重试按钮

## Files to Modify/Create

- Create: `cli/src/components/download-button.test.tsx`

## Steps

### Step 1: Test idle state — 显示"下载"文字和下载图标

### Step 2: Test loading state — 显示 spinner 和"获取中"，按钮 disabled

### Step 3: Test success state — 显示绿色 checkmark 和"成功"

### Step 4: Test error state — 显示"重试"，按钮可点击

### Step 5: Test click handler — 点击触发 onDownload 回调

### Step 6: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/components/download-button.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖四种按钮状态和点击交互
