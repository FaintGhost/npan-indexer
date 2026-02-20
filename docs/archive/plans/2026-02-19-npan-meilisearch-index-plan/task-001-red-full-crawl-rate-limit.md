# Task 001: 场景 1 红测（全量遍历与限流）

## Description

为“首次全量遍历并限流入索引”建立失败测试，先固定抓取流程的行为边界：分页遍历、目录队列推进、速率与并发上限。

## Execution Context

**Task Number**: 001 of 013  
**Phase**: Core Features  
**Prerequisites**: 已确认 `bdd-specs.md` 场景 1 内容。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 1：首次全量遍历并限流入索引  

## Files to Modify/Create

- Create: `tests/indexer/full-crawl-rate-limit.test.ts`
- Create: `tests/doubles/fake-npan-api.ts`
- Create: `tests/doubles/fake-meili-client.ts`

## Steps

### Step 1: Verify Scenario
- 确认测试断言覆盖“遍历完整性 + 限流阈值 + 输出统计”。

### Step 2: Implement Test (Red)
- 创建场景 1 对应测试。
- 使用 test doubles 隔离网络和 Meilisearch。
- 断言在实现前测试必须失败（缺少实现或行为不满足）。

### Step 3: Verify Red
- 仅运行该测试并确认失败原因是业务断言失败，而不是导入错误。

## Verification Commands

```bash
bun test tests/indexer/full-crawl-rate-limit.test.ts
```

## Success Criteria

- 测试稳定失败（Red）。
- 失败原因直接对应场景 1 行为缺失。
