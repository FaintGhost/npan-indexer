# Task 010: 场景 5 绿测（Meili 查询实现）

**depends-on**: task-009-red-meili-query-api

## Description

实现查询服务，直接使用 Meilisearch 提供检索结果与分页信息。

## Execution Context

**Task Number**: 010 of 013  
**Phase**: Integration  
**Prerequisites**: Task 009 红测已准备完成。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 5：搜索走 Meilisearch 而非平台弱检索  

## Files to Modify/Create

- Create: `src/search/query-service.ts`
- Create: `src/cli/search-local-index.ts`
- Modify: `tests/search/query-api.test.ts`

## Steps

### Step 1: Implement Logic (Green)
- 实现关键词搜索 + 过滤 + 排序 + 分页。
- 返回项包含 `file_id`、路径、类型等下载前置字段。

### Step 2: Verify Green
- 场景 5 测试通过。

### Step 3: Verify & Refactor
- 统一查询参数校验与错误映射。

## Verification Commands

```bash
bun test tests/search/query-api.test.ts
bun test
```

## Success Criteria

- 场景 5 测试通过。
- 查询结果满足业务过滤与排序要求。
