# Task 001: 编写索引设置优化测试 (Red)

**depends-on**: (none)

## Description

编写测试验证 `EnsureSettings` 传递给 Meilisearch 的 Settings 对象包含正确的优化配置：typo tolerance（文件扩展名禁用 typo）、中文 stop words、displayedAttributes（排除不需要展示的字段）、proximityPrecision 设为 byAttribute。

## Execution Context

**Task Number**: 1 of 5
**Phase**: Testing
**Prerequisites**: 理解当前 `EnsureSettings` 实现（`internal/search/meili_index.go:51-61`）

## BDD Scenario Reference

**Scenario**: 索引设置包含 typo tolerance 配置

```gherkin
Given 一个 mock IndexManager
When EnsureSettings 被调用
Then 传递给 UpdateSettingsWithContext 的 Settings 应满足：
  - TypoTolerance.Enabled == true
  - TypoTolerance.DisableOnAttributes 包含 "path_text"
  - TypoTolerance.DisableOnWords 包含常见文件扩展名（pdf, docx, xlsx, pptx, jpg, png, mp4, zip, rar, exe, apk, bin, iso）
  - StopWords 包含中文常用停用词（的, 了, 在, 是, 和, 就, 不, 都, 一, 一个, 上, 也, 到, 要, 会, 着, 没有, 好, 这）
  - DisplayedAttributes 不包含 "sha1", "in_trash", "is_deleted"
  - ProximityPrecision == ByAttribute
```

## Files to Modify/Create

- Create: `internal/search/meili_index_settings_test.go`

## Steps

### Step 1: 创建测试文件

创建 `internal/search/meili_index_settings_test.go`，实现一个 mock `meilisearch.IndexManager`，能捕获 `UpdateSettingsWithContext` 接收到的 `*meilisearch.Settings` 参数。

### Step 2: 编写测试用例

- `TestEnsureSettings_TypoTolerance` — 验证 TypoTolerance 设置：Enabled=true、DisableOnAttributes 包含 "path_text"、DisableOnWords 包含文件扩展名列表
- `TestEnsureSettings_StopWords` — 验证 StopWords 包含中文停用词列表
- `TestEnsureSettings_DisplayedAttributes` — 验证 DisplayedAttributes 已设置且不包含 "sha1"、"in_trash"、"is_deleted"
- `TestEnsureSettings_ProximityPrecision` — 验证 ProximityPrecision == `meilisearch.ByAttribute`

### Step 3: 验证测试失败 (Red)

- **Verification**: `go test ./internal/search/ -run TestEnsureSettings_ -v` → 应全部 FAIL

## Verification Commands

```bash
go test ./internal/search/ -run TestEnsureSettings_ -v
```

## Success Criteria

- 4 个测试用例均编写完成
- 所有测试在当前代码下 FAIL（Red 状态）
- Mock 正确捕获 Settings 参数
