# Task 012: 实现工具函数

**depends-on**: task-011

## Description

实现 formatBytes、formatTime、getFileIcon 纯函数，使 Task 011 测试通过。逻辑从现有 `web/app/index.html` 中迁移。

## Execution Context

**Task Number**: 012 of 046
**Phase**: Core Features (Green)
**Prerequisites**: Task 011 测试已编写（Red）

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 文件卡片正确展示信息

## Files to Modify/Create

- Create: `cli/src/lib/format.ts` — formatBytes, formatTime
- Create: `cli/src/lib/file-icon.ts` — getFileIcon（返回图标类型标识，不返回 SVG）

## Steps

### Step 1: Implement formatBytes — 参考现有 HTML 中 formatBytes 函数

### Step 2: Implement formatTime — 参考现有 HTML 中 formatTime 函数

### Step 3: Implement getFileIcon

- 输入文件名，返回图标分类对象（category, bgClass, textClass）
- 5 种分类：archive, installer, firmware, document, default
- 分类逻辑与现有 HTML 中 getFileIcon 一致

### Step 4: Verify tests PASS (Green)

## Verification Commands

```bash
cd cli && npx vitest run src/lib/format.test.ts src/lib/file-icon.test.ts
# Expected: PASS (Green)
```

## Success Criteria

- Task 011 所有测试通过
- 函数有完整 TypeScript 类型标注
