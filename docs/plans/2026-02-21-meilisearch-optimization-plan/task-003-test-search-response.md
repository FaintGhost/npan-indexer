# Task 003: 编写搜索响应优化测试 (Red)

**depends-on**: (none)

## Description

编写测试验证 Search 方法的请求中包含 AttributesToRetrieve（限定返回字段）和 AttributesToHighlight（对 name 字段高亮），并验证响应中包含高亮后的文件名。

## Execution Context

**Task Number**: 3 of 5
**Phase**: Testing
**Prerequisites**: 理解当前 `Search` 实现（`internal/search/meili_index.go:88-143`）和 `QueryResult` 结构（`internal/search/query_service.go`）

## BDD Scenario Reference

**Scenario**: 搜索请求包含高亮和字段限定参数

```gherkin
Given 一个 mock IndexManager 返回包含 _formatted 字段的搜索结果
When Search 被调用
Then 传递给 Meilisearch 的 SearchRequest 应满足：
  - AttributesToRetrieve 已设置（不包含 sha1, in_trash, is_deleted）
  - AttributesToHighlight 包含 "name"
  - HighlightPreTag == "<mark>"
  - HighlightPostTag == "</mark>"
And 返回的 QueryResult 应满足：
  - 每个 item 的 HighlightedName 包含 Meilisearch 返回的高亮文本
```

## Files to Modify/Create

- Create: `internal/search/meili_index_search_test.go`

## Steps

### Step 1: 创建测试文件

创建 `internal/search/meili_index_search_test.go`，实现 mock IndexManager，能捕获 Search 接收到的 SearchRequest 参数并返回带 `_formatted` 字段的响应。

### Step 2: 编写测试用例

- `TestSearch_RequestIncludesHighlightParams` — 验证 SearchRequest 中包含 AttributesToHighlight=["name"]、HighlightPreTag="<mark>"、HighlightPostTag="</mark>"
- `TestSearch_RequestIncludesAttributesToRetrieve` — 验证 SearchRequest 中 AttributesToRetrieve 已设置且不包含 sha1/in_trash/is_deleted
- `TestSearch_ResponseIncludesHighlightedName` — 验证返回的结果中 HighlightedName 字段包含 Meilisearch `_formatted.name` 的值

### Step 3: 验证测试失败 (Red)

- **Verification**: `go test ./internal/search/ -run TestSearch_ -v` → 应 FAIL

## Verification Commands

```bash
go test ./internal/search/ -run TestSearch_ -v
```

## Success Criteria

- 3 个测试用例编写完成
- 所有测试在当前代码下 FAIL（Red 状态）
