# Task 011: 测试工具函数（formatBytes, formatTime, getFileIcon）

**depends-on**: task-004

## Description

为纯工具函数创建失败测试用例。这些函数从现有 HTML JavaScript 中迁移而来。

## Execution Context

**Task Number**: 011 of 046
**Phase**: Core Features (Red)
**Prerequisites**: Task 004 测试基础设施已配置

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 1 - 文件卡片正确展示信息（大小格式化、时间格式化、扩展名图标）

## Files to Modify/Create

- Create: `cli/src/lib/format.test.ts`
- Create: `cli/src/lib/file-icon.test.ts`

## Steps

### Step 1: Test formatBytes

- `0` → `"-"`
- `undefined/null` → `"-"`
- `1024` → `"1 KB"`
- `1048576` → `"1 MB"`
- `1073741824` → `"1 GB"`
- `500` → `"500 B"`
- `1500` → `"1.5 KB"`
- `15360` → `"15 KB"`（>=10 时无小数）

### Step 2: Test formatTime

- `0` → `"-"`
- `undefined/null` → `"-"`
- Unix 秒级时间戳（如 `1700000000`）正确转换
- Unix 毫秒级时间戳（如 `1700000000000`）正确转换
- 格式为 `"YYYY-MM-DD HH:mm"`

### Step 3: Test getFileIcon

- `.zip/.rar/.7z/.tar/.gz` → amber 压缩包图标
- `.apk/.ipa/.exe/.dmg` 或名称含"安装包" → emerald 安装包图标
- `.bin/.iso/.img/.rom` 或名称含"固件" → purple 固件图标
- `.pdf/.doc/.docx/.txt/.md` → rose 文档图标
- 其他扩展名 → blue 通用文件图标
- 无扩展名 → 通用文件图标

### Step 4: Verify tests FAIL (Red)

## Verification Commands

```bash
cd cli && npx vitest run src/lib/format.test.ts src/lib/file-icon.test.ts
# Expected: FAIL (Red)
```

## Success Criteria

- 测试覆盖所有边界值和分类规则
- 测试因模块不存在而失败
