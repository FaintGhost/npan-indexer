# Task 002: 场景 1 绿测（全量遍历与限流实现）

**depends-on**: task-001-red-full-crawl-rate-limit

## Description

实现全量遍历任务主流程，并满足速率与并发限制，令场景 1 测试通过。

## Execution Context

**Task Number**: 002 of 013  
**Phase**: Core Features  
**Prerequisites**: Task 001 红测已就位并稳定失败。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 1：首次全量遍历并限流入索引  

## Files to Modify/Create

- Create: `src/indexer/full-crawl.ts`
- Create: `src/indexer/rate-limiter.ts`
- Modify: `tests/indexer/full-crawl-rate-limit.test.ts`

## Steps

### Step 1: Implement Logic (Green)
- 新建全量遍历入口与目录队列推进能力。
- 集成统一限流器，确保请求速率和并发可配置。
- 输出同步统计信息（处理数量、耗时、错误数）。

### Step 2: Verify Green
- 运行场景 1 测试并确保通过。

### Step 3: Verify & Refactor
- 抽离重复逻辑，保持测试通过。

## Verification Commands

```bash
bun test tests/indexer/full-crawl-rate-limit.test.ts
bun test
```

## Success Criteria

- 场景 1 测试转绿。
- 全量遍历与限流配置可被测试稳定验证。
