# Task 009: 场景 5 红测（Meili 查询接口）

## Description

为用户查询接口建立失败测试，确保搜索链路只依赖 Meilisearch。

## Execution Context

**Task Number**: 009 of 013  
**Phase**: Integration  
**Prerequisites**: 场景 5 过滤与排序规则已明确。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 5：搜索走 Meilisearch 而非平台弱检索  

## Files to Modify/Create

- Create: `tests/search/query-api.test.ts`
- Create: `tests/doubles/fake-query-client.ts`

## Steps

### Step 1: Verify Scenario
- 覆盖关键词、类型、更新时间过滤与稳定排序。

### Step 2: Implement Test (Red)
- 使用 fake query client，断言不得调用平台 `item/search`。
- 在实现前保证测试失败。

### Step 3: Verify Red
- 确认失败原因是查询实现缺失。

## Verification Commands

```bash
bun test tests/search/query-api.test.ts
```

## Success Criteria

- 场景 5 测试稳定失败。
