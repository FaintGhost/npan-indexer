# Task 012: 场景 6 绿测（下载 URL 代理实现）

**depends-on**: task-011-red-download-url-proxy

## Description

实现下载代理服务，按需向云盘请求 `download_url` 并返回给调用方。

## Execution Context

**Task Number**: 012 of 013  
**Phase**: Integration  
**Prerequisites**: Task 011 红测已存在。  

## BDD Scenario Reference

**Spec**: `./bdd-specs.md`  
**Scenario**: 场景 6：按需获取真实下载链接  

## Files to Modify/Create

- Create: `src/download/download-url-service.ts`
- Create: `src/cli/get-download-url.ts`
- Modify: `tests/download/download-url-proxy.test.ts`

## Steps

### Step 1: Implement Logic (Green)
- 实现按 `file_id` 调用云盘下载接口。
- 规范化错误输出（鉴权失败、文件不存在、限流等）。
- 保证临时 URL 不写入 Meili 文档。

### Step 2: Verify Green
- 场景 6 测试通过。

### Step 3: Verify & Refactor
- 抽离接口客户端与业务编排层，便于后续替换。

## Verification Commands

```bash
bun test tests/download/download-url-proxy.test.ts
bun test
```

## Success Criteria

- 场景 6 测试通过。
- 下载代理可稳定返回实时临时 URL。
