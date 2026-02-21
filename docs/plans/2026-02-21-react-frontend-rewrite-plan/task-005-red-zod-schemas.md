# Task 005: 测试 Zod Schema（搜索/下载/错误响应）

**depends-on**: task-004

## Description

为搜索响应、下载链接响应、错误响应创建 Zod schema 的失败测试用例。验证有效数据解析、无效数据拒绝、边界值处理。

## Execution Context

**Task Number**: 005 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施已配置

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 8 - 搜索 API 响应通过 schema 校验; 搜索 API 响应结构异常时优雅降级; 下载 URL API 响应通过 schema 校验; 错误响应通过 schema 校验

## Files to Modify/Create

- Create: `cli/src/schemas/search.test.ts`
- Create: `cli/src/schemas/download.test.ts`
- Create: `cli/src/schemas/error.test.ts`

## Steps

### Step 1: Create search schema test

- 测试有效搜索响应（包含 items 数组和 total）能被成功解析
- 测试 items 中每个元素包含必填字段（doc_id, source_id, name, size 等）
- 测试缺少 total 字段时 safeParse 返回失败
- 测试 highlighted_name 可选字段缺失时使用默认值
- 测试 items 为空数组时正常解析

### Step 2: Create download schema test

- 测试有效下载响应（file_id + download_url）能被解析
- 测试 download_url 为空字符串时解析失败
- 测试缺少 file_id 时解析失败

### Step 3: Create error schema test

- 测试有效错误响应（code + message）能被解析
- 测试 request_id 为可选字段

### Step 4: Verify tests FAIL (Red)

- 运行测试，确认因 schema 文件不存在而失败

## Verification Commands

```bash
cd cli && npx vitest run src/schemas/
# Expected: ALL FAIL (Red) — modules not found
```

## Success Criteria

- 所有测试因导入路径不存在而失败（Red 状态）
- 测试用例覆盖有效数据、无效数据、边界值
