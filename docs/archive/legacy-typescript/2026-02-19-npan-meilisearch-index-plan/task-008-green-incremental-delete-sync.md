# Task 008: 场景 4 绿测（增量与删除同步实现）

**depends-on**: task-007-red-incremental-delete-sync

## Description

实现增量同步作业与删除同步策略，保证状态推进与索引状态一致。

## Execution Context

**Task Number**: 008 of 013  
**Phase**: Core Features  
**Prerequisites**: Task 007 红测已完成。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 4：增量同步与删除同步  

## Files to Modify/Create

- Create: `src/indexer/incremental-sync.ts`
- Create: `src/indexer/sync-state-store.ts`
- Modify: `src/search/meili-index.ts`

## Steps

### Step 1: Implement Logic (Green)
- 实现按 `last_sync_time` 的增量拉取与比对。
- 实现删除同步（软删标记或硬删，需配置化）。
- 同步成功后原子推进同步游标。

### Step 2: Verify Green
- 运行场景 4 测试并通过。

### Step 3: Verify & Refactor
- 确保状态持久化与业务逻辑解耦。

## Verification Commands

```bash
bun test tests/indexer/incremental-delete-sync.test.ts
bun test
```

## Success Criteria

- 场景 4 测试通过。
- 增量与删除同步行为可重复执行且结果一致。
