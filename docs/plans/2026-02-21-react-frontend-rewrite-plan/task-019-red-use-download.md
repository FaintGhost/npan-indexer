# Task 019: 测试 useDownload Hook（获取链接、缓存、状态管理）

**depends-on**: task-010, task-004

## Description

为 useDownload 自定义 Hook 创建失败测试用例。

## Execution Context

**Task Number**: 019 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 010 API 客户端已实现

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 2 - 点击下载按钮获取下载链接; 下载按钮在请求期间显示加载状态; 下载链接获取成功后打开新标签页; 重复下载同一文件使用缓存链接; 下载链接获取失败显示重试按钮; 下载链接为空视为错误

## Files to Modify/Create

- Create: `cli/src/hooks/use-download.test.ts`

## Steps

### Step 1: Test download request

- 调用 download(42) → 发送 GET /api/v1/app/download-url?file_id=42

### Step 2: Test loading state

- 请求进行中 → status 为 "loading"

### Step 3: Test success state

- 请求成功 → status 为 "success"，url 不为空
- 1.5 秒后 status 回到 "idle"

### Step 4: Test cache hit

- 第二次调用 download(42) → 不发送请求，直接返回缓存 URL

### Step 5: Test error state

- API 返回 502 → status 为 "error"

### Step 6: Test empty URL is error

- download_url 为 "" → 视为错误

### Step 7: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/hooks/use-download.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖下载流程所有状态转换
