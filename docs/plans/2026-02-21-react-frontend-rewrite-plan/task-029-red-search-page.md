# Task 029: 测试搜索页面完整流程

**depends-on**: task-018, task-020, task-024, task-026, task-016, task-014, task-022, task-028

## Description

为搜索页面的完整用户流程创建集成测试。

## Execution Context

**Task Number**: 029 of 046
**Phase**: Integration (Red)
**Prerequisites**: 所有搜索相关 hooks 和组件已实现

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 初始空状态显示引导提示; 输入关键词后自动搜索并显示结果; 搜索无结果时显示空状态; 搜索 API 返回错误时显示错误状态; 点击清空按钮恢复初始状态; Feature 2 - 下载链接获取成功后打开新标签页

## Files to Modify/Create

- Create: `cli/src/routes/index.test.tsx`

## Steps

### Step 1: Test initial state — Hero 模式，显示引导提示

### Step 2: Test search flow — 输入关键词 → 骨架屏 → 结果卡片

### Step 3: Test no results — 搜索返回空 → 无结果状态

### Step 4: Test error — API 错误 → 错误状态

### Step 5: Test clear — 清空按钮 → 回到初始状态

### Step 6: Test download — 点击下载按钮 → 获取链接

### Step 7: Verify tests FAIL (Red) — SearchPage 组件不存在

## Verification Commands

```bash
cd cli && npx vitest run src/routes/index.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 集成测试覆盖搜索页面核心用户流程
