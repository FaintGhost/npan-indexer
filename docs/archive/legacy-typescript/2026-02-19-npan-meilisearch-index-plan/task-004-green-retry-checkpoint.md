# Task 004: 场景 2 绿测（重试与断点续跑实现）

**depends-on**: task-003-red-retry-checkpoint

## Description

实现重试退避与 checkpoint 持久化，使任务可恢复执行并通过场景 2 测试。

## Execution Context

**Task Number**: 004 of 013  
**Phase**: Core Features  
**Prerequisites**: Task 003 已就绪。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 2：限流/错误重试与断点续跑  

## Files to Modify/Create

- Create: `src/indexer/retry-policy.ts`
- Create: `src/indexer/checkpoint-store.ts`
- Modify: `src/indexer/full-crawl.ts`

## Steps

### Step 1: Implement Logic (Green)
- 增加指数退避 + 抖动策略。
- 增加 checkpoint 读写与恢复入口。
- 失败分支写入诊断日志并继续后续可处理队列。

### Step 2: Verify Green
- 场景 2 测试应通过。

### Step 3: Verify & Refactor
- 统一错误分类与日志字段，避免重复逻辑。

## Verification Commands

```bash
bun test tests/indexer/retry-checkpoint.test.ts
bun test
```

## Success Criteria

- 场景 2 测试通过。
- 中断恢复能力可在测试中复现并验证。
