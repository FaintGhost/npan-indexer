# Task 043: 测试可访问性（ARIA 属性、键盘导航、焦点管理）

**depends-on**: task-030, task-040

## Description

为应用的可访问性要求创建测试用例。

## Execution Context

**Task Number**: 043 of 046
**Phase**: Refinement (Red)
**Prerequisites**: 搜索页和管理页已实现

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 9 - 搜索结果区域为 ARIA live region; 搜索输入框有正确的 ARIA 标签; 下载按钮有描述性标签; 键盘可完整操作; 加载状态有 ARIA 反馈

## Files to Modify/Create

- Create: `cli/src/tests/accessibility.test.tsx`

## Steps

### Step 1: Test search results container has aria-live="polite"

### Step 2: Test search input has aria-label

### Step 3: Test download button has descriptive aria-label with file name

### Step 4: Test loading state has aria-busy="true"

### Step 5: Test skeleton has aria-hidden="true"

### Step 6: Test dialog focus trap

### Step 7: Test Tab navigation through interactive elements

### Step 8: Verify tests FAIL (Red) — ARIA attributes not yet added

## Verification Commands

```bash
cd cli && npx vitest run src/tests/accessibility.test.tsx
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖核心 ARIA 和键盘导航要求
