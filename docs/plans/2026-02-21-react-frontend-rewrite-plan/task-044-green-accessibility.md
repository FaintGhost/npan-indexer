# Task 044: 实现可访问性增强

**depends-on**: task-043

## Description

为所有组件添加 ARIA 属性和键盘导航支持，使 Task 043 测试通过。

## Execution Context

**Task Number**: 044 of 046
**Phase**: Refinement (Green)
**Prerequisites**: Task 043 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 9 - 可访问性所有场景

## Files to Modify/Create

- Modify: `cli/src/components/search-input.tsx` — aria-label
- Modify: `cli/src/components/file-card.tsx` — article role
- Modify: `cli/src/components/download-button.tsx` — aria-label with file name
- Modify: `cli/src/components/skeleton-card.tsx` — aria-hidden
- Modify: `cli/src/components/empty-state.tsx` — role
- Modify: `cli/src/components/api-key-dialog.tsx` — focus trap, aria-modal
- Modify: `cli/src/routes/index.tsx` — aria-live, aria-busy on results container

## Steps

### Step 1: Add aria-live="polite" to results container

### Step 2: Add aria-label to search input

### Step 3: Add descriptive aria-label to download buttons

### Step 4: Add aria-hidden to skeleton cards

### Step 5: Add aria-busy during loading

### Step 6: Ensure dialog has aria-modal and focus trap

### Step 7: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/tests/accessibility.test.tsx
# Expected: PASS (Green)
```

## Success Criteria

- Task 043 所有测试通过
- 使用 axe-core 或手动验证无明显 A11y 违规
