# Task 003: [TEST] `file_category` 索引契约 (RED)

**depends-on**: (none)

## Description

为 `file_category` 索引契约补充失败测试，锁定公开搜索分类必须依赖索引字段与 Meilisearch settings，而不是前端本地扩展名过滤。该任务只负责编写和验证失败测试，不实现生产代码。

## Execution Context

**Task Number**: 005 of 013
**Phase**: Foundation
**Prerequisites**: 已阅读 `internal/search/mapper.go`、`internal/search/meili_index.go` 与 `internal/models/models.go` 当前索引映射与 settings 结构

## BDD Scenario

```gherkin
Scenario: 文件分类筛选使用 file_category refinement
  Given 索引文档包含 file_category 字段并配置为 filterable
  When 用户选择 "文档" 分类筛选
  Then 搜索请求应携带对应 refinement
  And 结果总数应与筛选后的命中数一致
  And 页面不应再使用本地 items.filter 进行分类裁剪
```

**Spec Source**: `../2026-03-07-react-instantsearch-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `internal/search/mapper_test.go`
- Modify: `internal/search/meili_index_settings_test.go`

## Steps

### Step 1: Verify Scenario

- 确认场景聚焦“索引文档具备 `file_category` + settings 可筛选”这两层基础契约。
- 明确本任务不验证前端 refinement UI，只锁定后端索引基础能力。

### Step 2: Implement Test (Red)

- 在 mapper 测试中增加文件名到 `file_category` 的映射断言，覆盖文档、图片、视频、压缩包与其他分类。
- 在 settings 测试中增加 `FilterableAttributes` 对 `file_category` 的约束断言。
- 如公开搜索结果需要直接消费该字段，在 settings 测试中同步锁定 `DisplayedAttributes` 的公开边界。

### Step 3: Verify Red Failure

- 运行目标 Go 测试并确认失败。
- 失败原因应指向“缺少 `file_category` 字段或 settings 暴露不足”，而不是测试替身或外部服务问题。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/search -run 'FileCategory|MapFileToIndexDoc|EnsureSettings' -count=1
```

## Success Criteria

- 新增测试稳定处于 Red。
- 失败明确指向 `file_category` 索引契约缺失。
- 测试不依赖真实 Meilisearch 或外部网络。
