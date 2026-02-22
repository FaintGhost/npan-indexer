# Best Practices

## 1. 选择器策略

**优先级**（从高到低）：

1. `getByRole()` — 语义化，不依赖实现细节
2. `getByPlaceholder()` / `getByText()` — 用户可见文本
3. `getByTestId()` — 显式测试标识
4. `page.locator('css')` — 仅作为最后手段

```typescript
// 推荐
page.getByRole('button', { name: '启动同步' })
page.getByPlaceholder(/输入文件名关键词/)

// 避免
page.locator('.bg-blue-600.text-white.rounded-lg')  // Tailwind 类名会变
page.locator('#submit-btn')  // 魔法 ID
```

## 2. 等待策略

### 禁止使用 `waitForTimeout`

```typescript
// 错误：硬编码等待时间，CI 中不可靠
await page.waitForTimeout(500)

// 正确：等待 API 响应
const response = page.waitForResponse('**/api/v1/app/search*')
await searchInput.fill('query')
await response

// 正确：等待 DOM 状态
await expect(page.locator('article')).toHaveCount(30)
```

### 防抖输入的正确处理

```typescript
// 方案 A：waitForResponse（推荐）
const apiResponse = page.waitForResponse(r =>
  r.url().includes('/api/v1/app/search') && r.status() === 200
)
await searchInput.fill('test')
await apiResponse

// 方案 B：点击搜索按钮跳过防抖
await searchInput.fill('test')
await page.getByRole('button', { name: '搜索' }).click()
```

### 无限滚动

```typescript
// 滚动哨兵进入视口（IntersectionObserver 在真实浏览器中原生工作）
await page.locator('.h-2').last().scrollIntoViewIfNeeded()

// 等待新内容而非固定时间
await expect(page.locator('article')).toHaveCount(35, { timeout: 10_000 })
```

## 3. 认证处理

### 推荐：`addInitScript` 注入 localStorage

```typescript
// 在页面加载前就设置，无需 reload
await context.addInitScript(({ key, value }) => {
  localStorage.setItem(key, value)
}, { key: 'npan_admin_api_key', value: API_KEY })

await page.goto('/admin/')
// 首次加载就已认证，无闪烁
```

### 不推荐

```typescript
// 需要额外 reload，有闪烁
await page.goto('/admin/')
await page.evaluate(k => localStorage.setItem('npan_admin_api_key', k), API_KEY)
await page.reload()
```

## 4. Mock 策略

### 仅 mock 不可控的外部依赖

```typescript
// CI 中 NPA_TOKEN 为 dummy，下载 API 必然失败 → mock
await page.route('**/api/v1/app/download-url**', route =>
  route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({
      file_id: 1001,
      download_url: 'https://example.com/fake-download.pdf',
    }),
  })
)
```

### 不 mock 可控的内部服务

搜索、同步等走真实 Meilisearch/npan 后端，保证 E2E 真实性。

## 5. 数据播种

### Meilisearch 播种模式

```typescript
// beforeAll 中一次性播种
test.beforeAll(async () => {
  await seedMeilisearch()  // 插入 38 条文档 + 等待索引完成
})

test.afterAll(async () => {
  await clearMeilisearch()  // 清理测试数据
})
```

### 等待异步索引

```typescript
// Meilisearch 索引是异步的，必须轮询任务状态
async function waitForTask(taskUid: number) {
  while (true) {
    const task = await fetch(`${MEILI}/tasks/${taskUid}`)
    const { status } = await task.json()
    if (status === 'succeeded') return
    if (status === 'failed') throw new Error('seed failed')
    await sleep(200)
  }
}
```

## 6. window.open 拦截

```typescript
// 在测试开始时拦截 window.open
await page.addInitScript(() => {
  const calls: string[] = []
  ;(window as any).__openCalls = calls
  window.open = (url?: string | URL) => {
    if (url) calls.push(String(url))
    return null
  }
})

// 验证
await downloadButton.click()
const urls = await page.evaluate(() => (window as any).__openCalls)
expect(urls).toHaveLength(1)
expect(urls[0]).toContain('example.com')
```

## 7. CI 性能优化

| 配置 | 值 | 原因 |
|------|-----|------|
| browsers | Chromium only | 减少 50% 测试时间 |
| workers | 1 | 避免 Meilisearch 数据竞争 |
| retries | 1 (CI only) | 覆盖偶发网络抖动 |
| video | retain-on-failure | 不增加正常运行开销 |
| trace | on-first-retry | 仅重试时录制 |
| ipc | host | 防止 Chromium 共享内存不足崩溃 |

## 8. 常见陷阱

| 陷阱 | 正确做法 |
|------|---------|
| `waitForTimeout(300)` 处理防抖 | `waitForResponse()` |
| 动作后设置 `waitForEvent` | 先设置监听再触发动作 |
| `scrollTo(0, 99999)` 触发无限滚动 | `scrollIntoViewIfNeeded()` |
| Playwright 镜像版本与 npm 包不一致 | 严格匹配版本号 |
| Docker 中不加 `ipc: host` | 必须加，否则 Chromium 崩溃 |
| 用 CSS 类名定位元素 | 用 role/placeholder/text |
| `if: failure()` 上传 artifacts | 用 `if: ${{ !cancelled() }}` |
| 提交 `.auth/` 文件到 git | 加入 `.gitignore` |
| 在 `afterAll` 中不清理测试数据 | 始终清理，保证幂等性 |

## 9. 本地开发体验

```bash
# 启动 dev server + 打开交互式 UI
cd web && bun run e2e:ui

# 调试模式（步进执行）
cd web && bun run e2e:debug

# 仅运行搜索测试
cd web && bun run e2e -- --grep "搜索"

# 查看最新报告
cd web && bunx playwright show-report
```
