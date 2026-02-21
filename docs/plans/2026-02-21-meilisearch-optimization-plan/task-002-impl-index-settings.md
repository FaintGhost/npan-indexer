# Task 002: 实现索引设置优化 (Green)

**depends-on**: task-001

## Description

更新 `EnsureSettings` 方法，在现有的 searchableAttributes/filterableAttributes/sortableAttributes 基础上，添加 TypoTolerance、StopWords、DisplayedAttributes、ProximityPrecision 配置。

## Execution Context

**Task Number**: 2 of 5
**Phase**: Implementation
**Prerequisites**: Task 001 的测试已编写

## BDD Scenario Reference

**Scenario**: 同 Task 001 — 使测试从 Red 变为 Green

## Files to Modify/Create

- Modify: `internal/search/meili_index.go` — `EnsureSettings` 方法

## Steps

### Step 1: 更新 EnsureSettings

在 `EnsureSettings` 的 `meilisearch.Settings` 中添加以下字段：

- **TypoTolerance**: 设置 Enabled=true，DisableOnAttributes 包含 "path_text"，DisableOnWords 包含常见文件扩展名
- **StopWords**: 中文常用停用词列表
- **DisplayedAttributes**: 仅包含需要展示的字段（doc_id, source_id, type, name, path_text, parent_id, modified_at, created_at, size）
- **ProximityPrecision**: 设为 `meilisearch.ByAttribute`

### Step 2: 验证测试通过 (Green)

- **Verification**: `go test ./internal/search/ -run TestEnsureSettings_ -v` → 应全部 PASS

### Step 3: 运行完整测试套件

- **Verification**: `go test ./... -count=1` → 无回归

## Verification Commands

```bash
go test ./internal/search/ -run TestEnsureSettings_ -v
go test ./... -count=1
```

## Success Criteria

- EnsureSettings 包含所有优化设置
- Task 001 的全部测试 PASS
- 无其他测试回归
