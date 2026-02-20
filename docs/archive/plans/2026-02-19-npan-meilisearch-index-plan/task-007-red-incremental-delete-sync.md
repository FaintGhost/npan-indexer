# Task 007: 场景 4 红测（增量与删除同步）

## Description

为增量同步和删除同步建立失败测试，固定游标推进规则与删改行为。

## Execution Context

**Task Number**: 007 of 013  
**Phase**: Core Features  
**Prerequisites**: 场景 4 验收标准已明确。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 4：增量同步与删除同步  

## Files to Modify/Create

- Create: `tests/indexer/incremental-delete-sync.test.ts`
- Create: `tests/doubles/fake-sync-state-store.ts`

## Steps

### Step 1: Verify Scenario
- 覆盖新增、更新、删除三种变更类型。
- 覆盖“成功后推进 `last_sync_time`，失败不推进”。

### Step 2: Implement Test (Red)
- 使用 fake state store 与 fake API 构建时序数据。
- 在实现前确保测试失败。

### Step 3: Verify Red
- 确认失败原因和场景 4 一致。

## Verification Commands

```bash
bun test tests/indexer/incremental-delete-sync.test.ts
```

## Success Criteria

- 场景 4 测试稳定失败。
