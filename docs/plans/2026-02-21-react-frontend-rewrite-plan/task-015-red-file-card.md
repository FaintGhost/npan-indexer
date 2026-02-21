# Task 015: 测试文件卡片组件

**depends-on**: task-004, task-012

## Description

为 FileCard 组件创建失败测试用例。测试文件信息展示、高亮名称渲染、扩展名图标选择。

## Execution Context

**Task Number**: 015 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 012 工具函数已实现，Task 004 测试基础设施

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 文件卡片正确展示信息

## Files to Modify/Create

- Create: `cli/src/components/file-card.test.tsx`

## Steps

### Step 1: Test card renders file name with highlighting

- 传入 highlighted_name 包含 `<mark>` 标签
- 渲染后应显示高亮文字

### Step 2: Test card renders formatted size

- 传入 size=1048576 → 应显示 "1 MB"

### Step 3: Test card renders formatted date

- 传入 modified_at → 应显示格式化日期

### Step 4: Test card renders correct icon by extension

- `.bin` 文件 → 固件图标样式类
- `.zip` 文件 → 压缩包图标样式类

### Step 5: Test card includes download button

- 应包含一个"下载"按钮

### Step 6: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/components/file-card.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖卡片的所有展示要素
