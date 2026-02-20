# Task 005: 场景 3 红测（Meili 文档结构与可检索性）

## Description

建立 Meilisearch 文档结构与索引设置的失败测试，锁定主键、可检索字段、可过滤/排序字段。

## Execution Context

**Task Number**: 005 of 013  
**Phase**: Integration  
**Prerequisites**: 场景 3 验收字段已明确。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 3：Meilisearch 文档结构与可检索性  

## Files to Modify/Create

- Create: `tests/indexer/meili-schema-searchability.test.ts`
- Create: `tests/doubles/fake-meili-settings.ts`

## Steps

### Step 1: Verify Scenario
- 覆盖字段：`id/type/name/path_text/parent_id/modified_at/in_trash`。

### Step 2: Implement Test (Red)
- 使用 fake Meili client 校验设置调用与文档格式。
- 在实现前确保测试失败。

### Step 3: Verify Red
- 运行测试并确认失败来源是索引契约未满足。

## Verification Commands

```bash
bun test tests/indexer/meili-schema-searchability.test.ts
```

## Success Criteria

- 场景 3 测试稳定失败。
