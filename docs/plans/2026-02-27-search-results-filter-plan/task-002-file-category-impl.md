# Task 002: [IMPL] 扩展名分类规则与边界 (GREEN)

**depends-on**: task-002-file-category-test.md

## Description

实现集中式扩展名分类模块，提供统一分类函数，支持多段扩展名和默认归类。

## Execution Context

**Task Number**: 004 of 012  
**Phase**: Core Features  
**Prerequisites**: `task-002-file-category-test.md` 已完成并处于 Red

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

- Create: `web/src/lib/file-category.ts`
- Modify: `web/src/routes/index.lazy.tsx`（仅在接入阶段引用分类能力）

## Steps

### Step 1: Implement Logic (Green)
- 创建分类枚举与扩展名集合。
- 提供文件名到分类的纯函数，并处理多段后缀优先匹配。

### Step 2: Verify Green
- 运行分类测试确保通过。

### Step 3: Regression Check
- 运行受影响的搜索页相关测试，确认新增模块不破坏现有行为。

## Verification Commands

```bash
cd web && bun vitest run src/lib/file-category.test.ts
cd web && bun vitest run src/components/search-page.test.tsx
```

## Success Criteria

- 分类测试全部 Green。
- 分类实现为纯函数，可复用、可测试。
