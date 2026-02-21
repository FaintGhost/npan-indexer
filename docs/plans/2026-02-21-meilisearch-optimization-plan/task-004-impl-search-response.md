# Task 004: 实现搜索响应优化 (Green)

**depends-on**: task-003

## Description

更新 Search 方法的 SearchRequest 添加 AttributesToRetrieve 和 highlighting 参数。更新 IndexDocument 或 QueryResult 模型以携带高亮后的文件名。修改 hit 解析逻辑以提取 `_formatted.name` 字段。

## Execution Context

**Task Number**: 4 of 5
**Phase**: Implementation
**Prerequisites**: Task 003 的测试已编写

## BDD Scenario Reference

**Scenario**: 同 Task 003 — 使测试从 Red 变为 Green

## Files to Modify/Create

- Modify: `internal/search/meili_index.go` — `Search` 方法
- Modify: `internal/models/models.go` — `IndexDocument` 添加 HighlightedName 字段
- Modify: `internal/search/query_service.go` — 如需调整 QueryResult

## Steps

### Step 1: 扩展 IndexDocument 模型

在 `IndexDocument` 中添加 `HighlightedName string` 字段（json tag: `highlighted_name,omitempty`），用于存储搜索高亮后的文件名。

### Step 2: 更新 Search 方法

在 SearchRequest 中添加：
- `AttributesToRetrieve`: 仅请求需要的字段（doc_id, source_id, type, name, path_text, parent_id, modified_at, created_at, size）
- `AttributesToHighlight`: ["name"]
- `HighlightPreTag`: "<mark>"
- `HighlightPostTag`: "</mark>"

### Step 3: 解析 _formatted 数据

修改 hit 解析逻辑，从每个 hit 的 `_formatted.name` 字段中提取高亮文本，填充到 `HighlightedName`。需要处理 Meilisearch 返回的 raw JSON hit 结构。

### Step 4: 验证测试通过 (Green)

- **Verification**: `go test ./internal/search/ -run TestSearch_ -v` → 应全部 PASS

### Step 5: 运行完整测试套件

- **Verification**: `go test ./... -count=1` → 无回归

## Verification Commands

```bash
go test ./internal/search/ -run TestSearch_ -v
go test ./... -count=1
```

## Success Criteria

- SearchRequest 包含 highlighting 和 attributesToRetrieve 参数
- IndexDocument 包含 HighlightedName 字段
- _formatted.name 正确提取到 HighlightedName
- Task 003 的全部测试 PASS
- 无其他测试回归
