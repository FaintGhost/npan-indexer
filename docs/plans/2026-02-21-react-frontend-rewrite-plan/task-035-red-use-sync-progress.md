# Task 035: 测试 useSyncProgress Hook（轮询、启停）

**depends-on**: task-008, task-010, task-004

## Description

为 useSyncProgress 自定义 Hook 创建失败测试用例。

## Execution Context

**Task Number**: 035 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 008 Sync schema 已实现，Task 010 API 客户端已实现

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 6 - 同步运行中自动轮询进度; 同步完成后停止轮询; 同步出错时展示错误信息; 无同步进度记录时显示提示; 点击启动全量同步; 启动同步成功; 已有同步任务运行时启动返回冲突错误; 确认取消同步; 取消同步成功

## Files to Modify/Create

- Create: `cli/src/hooks/use-sync-progress.test.ts`

## Steps

### Step 1: Test initial load — 获取当前进度

### Step 2: Test polling — status="running" 时每 3 秒获取进度

### Step 3: Test stop polling — status="done" 时停止

### Step 4: Test start sync — 发送 POST 请求

### Step 5: Test start sync conflict — 409 错误处理

### Step 6: Test cancel sync — 发送 POST cancel 请求

### Step 7: Test 404 — 无进度记录

### Step 8: Test cleanup — 组件卸载时停止轮询

### Step 9: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-sync-progress.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖同步管理的完整流程
