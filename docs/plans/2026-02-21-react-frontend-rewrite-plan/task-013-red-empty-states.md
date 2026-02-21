# Task 013: 测试骨架屏与空状态组件

**depends-on**: task-004, task-002

## Description

为骨架屏卡片、初始空状态、无结果状态、错误状态组件创建失败测试用例。

## Execution Context

**Task Number**: 013 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施，Task 002 Tailwind 配置

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 初始空状态显示引导提示; 搜索无结果时显示空状态; 搜索 API 返回错误时显示错误状态; Feature 3 - 首次搜索从 Hero 切换时显示骨架屏

## Files to Modify/Create

- Create: `cli/src/components/empty-state.test.tsx`
- Create: `cli/src/components/skeleton-card.test.tsx`

## Steps

### Step 1: Test InitialState component

- 渲染后应显示"等待探索"标题
- 应包含搜索图标
- 应包含引导描述文字

### Step 2: Test NoResultsState component

- 渲染后应显示"未找到相关文件"标题

### Step 3: Test ErrorState component

- 渲染后应显示"加载出错了"标题
- 应使用 rose/红色主题

### Step 4: Test SkeletonCard component

- 渲染 5 个骨架卡片
- 每个骨架应有 pulse 动画类
- 应设置 aria-hidden="true"

### Step 5: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/components/empty-state.test.tsx src/components/skeleton-card.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖三种空状态和骨架屏
