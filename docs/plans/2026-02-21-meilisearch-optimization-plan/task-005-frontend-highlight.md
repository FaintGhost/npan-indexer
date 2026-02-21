# Task 005: 前端渲染搜索高亮

**depends-on**: task-004

## Description

更新 Web 搜索页面（`web/app/index.html`）的 `cardHTML` 函数，优先使用 API 返回的 `highlighted_name` 字段渲染文件名（包含 `<mark>` 标签高亮），如果没有则回退到普通 `name`。同时添加 `<mark>` 标签的 CSS 样式。

## Execution Context

**Task Number**: 5 of 5
**Phase**: Refinement
**Prerequisites**: Task 004 已实现，API 返回 highlighted_name 字段

## BDD Scenario Reference

**Scenario**: 前端显示高亮搜索结果

```gherkin
Given API 返回的搜索结果包含 highlighted_name 字段（内含 <mark> 标签）
When 前端渲染搜索结果卡片
Then 文件名区域应使用 highlighted_name 的 HTML 内容
  And <mark> 标签应有明显的高亮样式（如黄色背景）
  And title 属性仍使用原始 name（不含 HTML 标签）
```

## Files to Modify/Create

- Modify: `web/app/index.html` — `cardHTML` 函数和 CSS 样式

## Steps

### Step 1: 添加 mark 标签样式

在 `<style>` 区域添加 `mark` 元素的样式：圆角、背景色（如 amber/yellow），使高亮词汇在视觉上突出。

### Step 2: 更新 cardHTML 函数

修改 `cardHTML` 函数中文件名 `<h3>` 的渲染逻辑：
- 如果 `item.highlighted_name` 存在，使用 `innerHTML` 直接渲染（因为包含 `<mark>` 标签）
- `title` 属性保持使用原始 `name`（经 `escapeHTML` 处理）
- 如果 `highlighted_name` 不存在，回退到原来的 `escapeHTML(name)` 渲染

### Step 3: 手动验证

启动服务后在浏览器中搜索，确认搜索关键词在文件名中被高亮显示。

## Verification Commands

```bash
# 启动服务
go run ./cmd/server

# 浏览器访问
# http://127.0.0.1:1323/app
# 搜索任意关键词，确认高亮效果
```

## Success Criteria

- 搜索结果中匹配的关键词以高亮样式显示
- title 属性显示原始文件名
- 无高亮数据时正常回退显示
