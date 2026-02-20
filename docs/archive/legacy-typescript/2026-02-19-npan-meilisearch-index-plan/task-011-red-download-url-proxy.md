# Task 011: 场景 6 红测（下载 URL 代理）

## Description

为按 `file_id` 动态换取真实下载链接建立失败测试，验证临时 URL 不落索引。

## Execution Context

**Task Number**: 011 of 013  
**Phase**: Integration  
**Prerequisites**: 场景 6 验收规则明确。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 6：按需获取真实下载链接  

## Files to Modify/Create

- Create: `tests/download/download-url-proxy.test.ts`
- Create: `tests/doubles/fake-download-api.ts`

## Steps

### Step 1: Verify Scenario
- 覆盖成功返回临时 URL、失败返回可诊断错误、索引不存 URL。

### Step 2: Implement Test (Red)
- 使用 fake 下载 API 模拟正常和异常路径。
- 在实现前确保测试失败。

### Step 3: Verify Red
- 失败原因应对应下载代理逻辑缺失。

## Verification Commands

```bash
bun test tests/download/download-url-proxy.test.ts
```

## Success Criteria

- 场景 6 测试稳定失败。
