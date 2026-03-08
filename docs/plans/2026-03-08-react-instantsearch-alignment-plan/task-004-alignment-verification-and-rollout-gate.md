# Task 004: 结果对比与回滚闸门验证

**depends-on**: task-001-public-search-as-you-type-impl.md, task-002-query-normalizer-impl.md, task-003-public-default-filters-impl.md

## Description

执行 public 与 legacy 搜索链路的结果对比、回归验证与发布/回滚闸门确认，确保达到“官方行为 + 关键对齐”后再考虑默认开启 public 搜索。

## Execution Context

**Task Number**: 007 of 007
**Phase**: Verification and Rollout Gate
**Prerequisites**: public 输入语义、query adapter 与默认过滤基线均已完成并通过对应 Green 任务

## BDD Scenario

```gherkin
Scenario: public 与 legacy 结果对比形成可发布结论
  Given 已准备代表性查询集，覆盖普通查询、扩展名查询、版本号查询和多词组合查询
  When 分别执行 public 搜索与 legacy AppSearch
  Then 应记录每个查询的命中总数、前 10 条结果、高亮输出、空态与错误态
  And 应将差异标记为可接受、待解释或阻塞级差异

Scenario: 阻塞级差异会阻止默认开启 public 搜索
  Given 已完成 public 与 legacy 的结果对比
  When 存在 folder 或 deleted 或 trash 泄漏
  Or 存在关键 preprocess 查询 legacy 有结果而 public 无结果
  Or public 仍然不是 search-as-you-type
  Then 本次发布不得默认开启 public 搜索
  And 必须保留 legacy fallback 作为主链路或回退链路

Scenario: 达到阻塞级差异或故障门槛时可立即回滚到 AppSearch
  Given public 搜索已在预发或线上灰度开启
  And 已出现阻塞级差异或不可接受故障
  When 运维关闭 instantsearchEnabled 开关
  Then 前端应切回 AppSearch 搜索链路
  And 下载与页面其他功能保持可用
  And 用户无需变更访问入口
```

**Spec Source**: `../2026-03-08-react-instantsearch-alignment-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `web/e2e/tests/search.spec.ts`
- Modify: `web/e2e/pages/search-page.ts`
- Modify: `tasks/todo.md`（回填执行结果与发布建议）

## Steps

### Step 1: Build Comparison Corpus

- 准备代表性查询集，至少覆盖普通查询、扩展名查询、版本号查询、多词组合查询。
- 明确每个查询要对比的维度：命中总数、前 10 条结果、高亮输出、空态与错误态。

### Step 2: Run Focused Regression

- 执行 public 输入即搜、query adapter、默认过滤 / refinement 的单测与页面回归。
- 通过 E2E 或浏览器级断言确认 public 模式下 `/multi-search` 行为、legacy fallback 与回滚路径仍可工作。

### Step 3: Evaluate Rollout Gate

- 将 public vs legacy 的差异标记为可接受、待解释或阻塞级差异。
- 若存在 folder / deleted / trash 泄漏、关键查询召回缺失或 search-as-you-type 回退，则不得默认开启 public 搜索。
- 保留 `instantsearchEnabled` 作为发布与回滚闸门，验证关闭开关后搜索与下载仍走 legacy 链路。

### Step 4: Record Sign-off

- 在 `tasks/todo.md` 回填对比结果、阻塞级差异结论、灰度建议与回滚条件。
- 明确本轮是否达到默认开启 public 搜索的门槛。

## Verification Commands

```bash
cd web && bun vitest run src/components/search-page.test.tsx src/components/search-filters.test.tsx src/lib/meili-search-client.test.ts src/lib/search-query-normalizer.test.ts
cd web && bun vitest run
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120
./tests/smoke/smoke_test.sh
docker compose -f docker-compose.ci.yml --profile e2e run --rm playwright --grep search
docker compose -f docker-compose.ci.yml --profile e2e down --volumes
git diff --check
```

## Success Criteria

- 已完成 public vs legacy 的代表性查询结果对比，并有明确差异结论。
- 已验证 search-as-you-type、query adapter、默认过滤与 refinement 叠加在浏览器链路中成立。
- 已定义 `instantsearchEnabled` 的灰度启用与快速回滚条件。
- 默认启用前不存在未解释的阻塞级差异。
