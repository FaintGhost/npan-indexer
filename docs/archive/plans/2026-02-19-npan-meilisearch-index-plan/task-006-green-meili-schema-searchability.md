# Task 006: 场景 3 绿测（Meili 写入与索引设置实现）

**depends-on**: task-005-red-meili-schema-searchability

## Description

实现 Meilisearch 索引初始化与文档 upsert，满足场景 3 的搜索/过滤/排序需求。

## Execution Context

**Task Number**: 006 of 013  
**Phase**: Integration  
**Prerequisites**: Task 005 红测完成。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 3：Meilisearch 文档结构与可检索性  

## Files to Modify/Create

- Create: `src/search/meili-index.ts`
- Create: `src/search/meili-mapper.ts`
- Modify: `src/indexer/full-crawl.ts`

## Steps

### Step 1: Implement Logic (Green)
- 定义文档主键策略（确保 `id + type` 全局唯一）。
- 配置 searchable/filterable/sortable 属性。
- 实现批量 upsert 与错误处理。

### Step 2: Verify Green
- 场景 3 测试应通过。

### Step 3: Verify & Refactor
- 校验映射层与索引层职责边界清晰。

## Verification Commands

```bash
bun test tests/indexer/meili-schema-searchability.test.ts
bun test
```

## Success Criteria

- 场景 3 测试通过。
- 文档结构与索引设置可重复初始化。
