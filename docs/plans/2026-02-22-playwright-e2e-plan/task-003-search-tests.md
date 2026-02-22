# Task 003: 搜索流程 E2E 测试

**depends-on**: Task 001

## Objective

编写搜索页面的 E2E 测试用例，覆盖 BDD specs Feature 1 中的所有搜索场景。

## BDD Scenarios Covered

- Scenario: 初始状态显示欢迎界面
- Scenario: 输入关键词后防抖触发搜索
- Scenario: 点击搜索按钮立即搜索（跳过防抖）
- Scenario: 按 Enter 键立即搜索
- Scenario: 无结果时显示空状态
- Scenario: 清空搜索恢复初始状态
- Scenario: 无限滚动加载更多
- Scenario: Cmd/Ctrl+K 聚焦搜索框
- Scenario: 视图模式切换动画

## Files to Create/Modify

| File | Action |
|------|--------|
| `web/e2e/tests/search.spec.ts` | 新建 |

## Steps

### 1. 创建 search.spec.ts 基本结构

- 导入 `test`、`expect` 从 `../fixtures/auth`
- 导入 `SearchPage` 从 `../pages/search-page`
- 导入 `seedMeilisearch`、`clearMeilisearch` 从 `../fixtures/seed`
- `test.beforeAll`: 调用 `seedMeilisearch()` 播种 38 条文档
- `test.afterAll`: 调用 `clearMeilisearch()` 清理

### 2. 编写「初始状态显示欢迎界面」测试

- 访问 `/`
- 断言 hero 模式 (`.mode-hero`) 可见
- 断言 "Npan Search" 标题可见
- 断言状态文本包含 "随时准备为您检索文件"
- 断言 `resultArticles` 数量为 0

### 3. 编写「输入关键词后防抖触发搜索」测试

- 使用 `page.waitForResponse()` 监听 `/api/v1/app/search?query=quarterly`
- 在搜索框输入 "quarterly"
- 等待 API 响应返回
- 断言页面切换到 docked 模式 (`.mode-docked`)
- 断言至少有 1 条结果
- 断言状态文本匹配 "已加载 N 个文件"

### 4. 编写「点击搜索按钮立即搜索」测试

- 在搜索框输入 "project"
- 使用 `searchImmediate()` 点击搜索按钮
- 断言结果列表至少有 1 条

### 5. 编写「按 Enter 键立即搜索」测试

- 在搜索框输入 "design"
- 按 Enter 键（`searchInput.press('Enter')`）
- 等待 API 响应
- 断言至少有 1 条结果

### 6. 编写「无结果时显示空状态」测试

- 搜索 "xyzzy-nonexistent-99999"
- 等待搜索完成
- 断言页面显示 "未找到相关文件"
- 断言结果列表为空

### 7. 编写「清空搜索恢复初始状态」测试

- 先搜索 "test" 并等待结果
- 点击清空按钮
- 断言搜索框为空
- 断言恢复 hero 模式
- 断言状态文本显示 "随时准备为您检索文件"

### 8. 编写「无限滚动加载更多」测试

- 搜索 "test-file"（应匹配 35 条批量文档）
- 断言首页加载 30 条结果
- 滚动到哨兵元素（`scrollToLoadMore()`）
- 等待第二页 API 响应（`page=2`）
- 断言结果总数为 35

### 9. 编写「Cmd/Ctrl+K 聚焦搜索框」测试

- 按下 `Meta+K`（Mac）或 `Control+K`
- 断言搜索框获得焦点（`toBeFocused()`）

### 10. 编写「视图模式切换」测试

- 断言初始 `.mode-hero`
- 输入搜索词，等待搜索完成
- 断言切换到 `.mode-docked`
- 清空搜索框
- 断言恢复 `.mode-hero`

## Verification

```bash
cd web && bunx playwright test search.spec.ts --list
# 应列出约 9 个搜索测试用例
```
