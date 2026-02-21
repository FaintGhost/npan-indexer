# Task 016: 实现文件卡片组件

**depends-on**: task-015

## Description

实现 FileCard 组件，使 Task 015 测试通过。组件接收 IndexDocument 类型 props。

## Execution Context

**Task Number**: 016 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 015 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 文件卡片正确展示信息

## Files to Modify/Create

- Create: `cli/src/components/file-card.tsx`

## Steps

### Step 1: Implement FileCard

- Props: IndexDocument 类型 + onDownload 回调
- 使用 dangerouslySetInnerHTML 渲染 highlighted_name（后端已转义）
- 使用 formatBytes 和 formatTime 格式化大小和日期
- 使用 getFileIcon 获取图标分类
- 包含下载按钮（点击触发 onDownload）
- 使用 React.memo 优化，以 source_id 为 key 时避免不必要重渲染

### Step 2: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/components/file-card.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 015 所有测试通过
- 组件样式匹配现有 HTML 中 cardHTML 函数的输出
