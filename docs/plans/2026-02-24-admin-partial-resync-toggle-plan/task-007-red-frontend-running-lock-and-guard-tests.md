# Task 007: RED frontend running lock and force-rebuild guard tests

## Description

先写失败测试约束风险交互：运行中禁用操作、`force_rebuild` 与局部补同步互斥。

## Execution Context

**Task Number**: 007 of 011  
**Phase**: Testing (Red)  
**Prerequisites**: 无

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `运行中禁止修改目录选择和拉取目录详情`、`运行中禁止重复发起局部补同步`、`force_rebuild 与局部补同步互斥`

## Files to Modify/Create

- Modify: `web/src/components/admin-page.test.tsx`

## Steps

### Step 1: Verify Scenario

- 明确三条交互防线：running lock、禁止重复提交、force_rebuild 互斥。

### Step 2: Implement Test (Red)

- 新增用例：
  - `status=running` 时 toggle 与 inspect 按钮禁用
  - `status=running` 时“启动同步”禁用/文案锁定
  - 有 scoped selection 且打开 force_rebuild 时阻止请求发送并提示错误

### Step 3: Verify Failure

- 在实现前运行并确认失败。

## Verification Commands

```bash
cd web && bun vitest run src/components/admin-page.test.tsx
```

## Success Criteria

- 红测能准确暴露缺失的防护行为。
- 用例不依赖真实后端。
