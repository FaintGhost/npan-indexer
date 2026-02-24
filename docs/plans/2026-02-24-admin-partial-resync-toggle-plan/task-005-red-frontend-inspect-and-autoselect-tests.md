# Task 005: RED frontend inspect decoupling and auto-select tests

**depends-on**: task-002-green-backend-inspect-roots-api-and-contract.md

## Description

先写前端失败测试，锁定“拉取目录详情不触发同步”以及“新目录默认自动勾选”的交互行为。

## Execution Context

**Task Number**: 005 of 011  
**Phase**: Testing (Red)  
**Prerequisites**: inspect roots API 契约已存在

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `拉取目录详情只更新列表，不启动同步`、`批量拉取目录详情部分成功`

## Files to Modify/Create

- Modify: `web/src/components/admin-page.test.tsx`
- Modify: `web/src/hooks/use-sync-progress.test.ts`（若新增 inspect hook）

## Steps

### Step 1: Verify Scenario

- 对照 BDD，明确“拉取”和“同步”是两个独立动作。

### Step 2: Implement Test (Red)

- 新增用例：
  - 点击“拉取目录详情”只调用 inspect API，不调用 `/api/v1/admin/sync`
  - inspect 返回成功项后对应目录在 UI 中默认处于勾选态
  - 部分失败时成功项仍可见，错误消息可见

### Step 3: Verify Failure

- 在未实现 UI 改造前确认用例失败。

## Verification Commands

```bash
cd web && bun vitest run src/components/admin-page.test.tsx src/hooks/use-sync-progress.test.ts
```

## Success Criteria

- 红测稳定失败并准确表达行为缺口。
- 测试不依赖真实网络（MSW mock）。
