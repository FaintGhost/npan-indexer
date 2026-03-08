# Task 003: [IMPL] `file_category` 索引契约 (GREEN)

**depends-on**: task-003-file-category-index-test.md

## Description

实现 `file_category` 索引字段与相关 Meilisearch settings 收敛，为后续 InstantSearch refinement 提供真实服务端筛选基础。

## Execution Context

**Task Number**: 006 of 013
**Phase**: Foundation
**Prerequisites**: `task-003-file-category-index-test.md` 已完成且处于 Red

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

- Modify: `internal/models/models.go`
- Modify: `internal/search/mapper.go`
- Modify: `internal/search/meili_index.go`
- Modify: `internal/search/mapper_test.go`
- Modify: `internal/search/meili_index_settings_test.go`
- Modify: `web/e2e/fixtures/seed.ts`

## Steps

### Step 1: Add Index Field Contract

- 在 `IndexDocument` 中引入 `file_category`，并明确其枚举值与前端 refinement 值域保持一致。
- 保证该字段属于公开搜索页允许暴露的最小必要数据集合。

### Step 2: Update Mapping and Settings

- 在索引映射阶段根据文件名分类结果写入 `file_category`。
- 在 Meilisearch settings 中将 `file_category` 加入 `FilterableAttributes`，并按公开搜索页渲染需要决定是否加入 `DisplayedAttributes`。

### Step 3: Align Seed Data

- 更新 E2E 种子文档，确保至少覆盖多种 `file_category` 命中，供后续 refinement 与结果回归复用。
- 保持种子数据与公开搜索首批分类模型一致，避免测试名称与索引值漂移。

### Step 4: Verify Green

- 运行 task-003 新增测试并确认通过。
- 回归 `internal/search` 相关测试，确认索引 settings 与映射逻辑无回归。

## Verification Commands

```bash
GOCACHE=/tmp/go-build go test ./internal/search -run 'FileCategory|MapFileToIndexDoc|EnsureSettings' -count=1
GOCACHE=/tmp/go-build go test ./internal/search -count=1
git diff --check
```

## Success Criteria

- `IndexDocument` 已具备 `file_category`。
- Meilisearch settings 已声明 `file_category` 可筛选。
- task-003 新增测试通过。
- E2E 种子数据可支撑后续分类 refinement 验证。
