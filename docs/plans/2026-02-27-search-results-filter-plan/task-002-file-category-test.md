# Task 002: [TEST] 扩展名分类规则与边界 (RED)

**depends-on**: (none)

## Description

新增文件分类规则单测，覆盖文档/图片/视频/压缩包/其他分类及多段扩展名边界，为过滤逻辑提供可验证基础。

## Execution Context

**Task Number**: 003 of 012  
**Phase**: Core Features  
**Prerequisites**: 测试环境可运行 Vitest

## BDD Scenario

```gherkin
Scenario: 分类规则正确识别常见扩展名
  Given 文件名分别为 report.pdf、photo.jpg、demo.mp4、backup.tar.gz、README
  When 系统执行扩展名分类
  Then report.pdf 属于 doc
  And photo.jpg 属于 image
  And demo.mp4 属于 video
  And backup.tar.gz 属于 archive
  And README 属于 other
```

**Spec Source**: `../2026-02-27-search-results-filter-design/bdd-specs.md`

## Files to Modify/Create

- Create: `web/src/lib/file-category.test.ts`

## Steps

### Step 1: Verify Scenario
- 确认设计文档包含分类规则场景。

### Step 2: Implement Test (Red)
- 新建分类单测，覆盖：大小写、无扩展名、未知扩展名、多段扩展名。
- 不依赖真实网络或外部服务。

### Step 3: Verify Red Failure
- 执行目标测试并确认失败（缺少实现或行为不符）。

## Verification Commands

```bash
cd web && bun vitest run src/lib/file-category.test.ts
```

## Success Criteria

- 分类规则测试失败且失败原因清晰（Red）。
- 测试独立可重复运行。
