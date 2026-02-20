# Task 003: 场景 2 红测（重试与断点续跑）

## Description

为“429/5xx 重试 + 断点续跑”建立失败测试，约束退避策略和 checkpoint 恢复行为。

## Execution Context

**Task Number**: 003 of 013  
**Phase**: Core Features  
**Prerequisites**: 场景 2 验收标准已明确。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 2：限流/错误重试与断点续跑  

## Files to Modify/Create

- Create: `tests/indexer/retry-checkpoint.test.ts`
- Create: `tests/doubles/fake-checkpoint-store.ts`

## Steps

### Step 1: Verify Scenario
- 明确断言：429/5xx 触发指数退避、超过重试上限后记录失败、重启从 checkpoint 继续。

### Step 2: Implement Test (Red)
- 使用 fake API 注入连续失败与恢复响应。
- 使用 fake checkpoint store 验证恢复点读写。

### Step 3: Verify Red
- 运行测试并确认失败原因对应场景 2 行为缺失。

## Verification Commands

```bash
bun test tests/indexer/retry-checkpoint.test.ts
```

## Success Criteria

- 场景 2 测试稳定失败。
- 失败信息能明确指向重试或恢复逻辑缺失。
