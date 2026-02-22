import { test, expect } from '../fixtures/auth'
import { SearchPage } from '../pages/search-page'
import { seedMeilisearch, clearMeilisearch } from '../fixtures/seed'

test.describe('搜索流程', () => {
  let searchPage: SearchPage

  test.beforeAll(async () => {
    await seedMeilisearch()
  })

  test.afterAll(async () => {
    await clearMeilisearch()
  })

  test.beforeEach(async ({ page }) => {
    searchPage = new SearchPage(page)
    await searchPage.goto()
  })

  // Test 1: 初始状态显示欢迎界面
  test('初始状态显示欢迎界面', async () => {
    // Assert hero mode visible
    await expect(searchPage.heroMode).toBeVisible()
    // Assert title "Npan Search" visible
    await expect(searchPage.page.getByRole('heading', { name: 'Npan Search' })).toBeVisible()
    // Assert status text "随时准备为您检索文件"
    await expect(searchPage.statusText).toHaveText('随时准备为您检索文件')
    // Assert no results
    await expect(searchPage.resultArticles).toHaveCount(0)
  })

  // Test 2: 输入关键词后防抖触发搜索
  test('输入关键词后防抖触发搜索', async () => {
    const response = await searchPage.search('quarterly')
    expect(response.status()).toBe(200)
    // Should switch to docked mode
    await expect(searchPage.dockedMode).toBeVisible()
    // Should have at least 1 result
    await searchPage.waitForResults()
    const count = await searchPage.getResultCount()
    expect(count).toBeGreaterThanOrEqual(1)
    // Status text should show loaded count
    await expect(searchPage.statusText).toContainText('已加载')
  })

  // Test 3: 点击搜索按钮立即搜索
  test('点击搜索按钮立即搜索（跳过防抖）', async () => {
    const response = await searchPage.searchImmediate('project')
    expect(response.status()).toBe(200)
    await searchPage.waitForResults()
    const count = await searchPage.getResultCount()
    expect(count).toBeGreaterThanOrEqual(1)
  })

  // Test 4: 按 Enter 键立即搜索
  test('按 Enter 键立即搜索', async () => {
    const responsePromise = searchPage.page.waitForResponse(
      r => r.url().includes('/api/v1/app/search') && r.status() === 200
    )
    await searchPage.searchInput.fill('design')
    await searchPage.searchInput.press('Enter')
    const response = await responsePromise
    expect(response.status()).toBe(200)
    await searchPage.waitForResults()
    const count = await searchPage.getResultCount()
    expect(count).toBeGreaterThanOrEqual(1)
  })

  // Test 5: 无结果时显示空状态
  test('无结果时显示空状态', async () => {
    await searchPage.search('xyzzy-nonexistent-99999')
    // Wait for loading to finish
    await expect(searchPage.page.getByRole('heading', { name: '未找到相关文件' })).toBeVisible()
    await expect(searchPage.resultArticles).toHaveCount(0)
  })

  // Test 6: 清空搜索恢复初始状态
  test('清空搜索恢复初始状态', async () => {
    // First search
    await searchPage.search('test')
    await searchPage.waitForResults()
    // Clear
    await searchPage.clearButton.click()
    // Assert reset
    await expect(searchPage.searchInput).toHaveValue('')
    await expect(searchPage.heroMode).toBeVisible()
    await expect(searchPage.statusText).toHaveText('随时准备为您检索文件')
  })

  // Test 7: 无限滚动加载更多
  test('无限滚动加载更多', async () => {
    // Search for "test-file" should match 35 bulk docs
    await searchPage.search('test-file')
    await searchPage.waitForResults()
    // First page should load 30 results
    await expect(searchPage.resultArticles).toHaveCount(30, { timeout: 10_000 })
    // Scroll to sentinel to trigger infinite scroll
    const secondPageResponse = searchPage.page.waitForResponse(
      r => r.url().includes('/api/v1/app/search') && r.url().includes('page=2') && r.status() === 200
    )
    await searchPage.scrollToLoadMore()
    await secondPageResponse
    // Total should be 35
    await expect(searchPage.resultArticles).toHaveCount(35, { timeout: 10_000 })
  })

  // Test 8: Cmd/Ctrl+K 聚焦搜索框
  test('Cmd/Ctrl+K 聚焦搜索框', async () => {
    // Click somewhere else first to unfocus
    await searchPage.page.click('body')
    // Press keyboard shortcut
    const modifier = process.platform === 'darwin' ? 'Meta' : 'Control'
    await searchPage.page.keyboard.press(`${modifier}+k`)
    // Assert search input is focused
    await expect(searchPage.searchInput).toBeFocused()
  })

  // Test 9: 视图模式切换
  test('视图模式切换', async () => {
    // Initially hero
    await expect(searchPage.heroMode).toBeVisible()
    // Search triggers docked
    await searchPage.search('test')
    await searchPage.waitForResults()
    await expect(searchPage.dockedMode).toBeVisible()
    // Clear returns to hero
    await searchPage.clearButton.click()
    await expect(searchPage.heroMode).toBeVisible()
  })
})

test.describe('下载流程', () => {
  let searchPage: SearchPage

  test.beforeAll(async () => {
    await seedMeilisearch()
  })

  test.afterAll(async () => {
    await clearMeilisearch()
  })

  test.beforeEach(async ({ page }) => {
    searchPage = new SearchPage(page)

    // Intercept window.open
    await page.addInitScript(() => {
      const calls: string[] = []
      ;(window as any).__openCalls = calls
      window.open = (url?: string | URL) => {
        if (url) calls.push(String(url))
        return null
      }
    })

    // Mock download API - success by default
    await page.route('**/api/v1/app/download-url**', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          file_id: 1001,
          download_url: 'https://example.com/fake-download.pdf',
        }),
      }),
    )

    await searchPage.goto()
    // Search and get results first
    await searchPage.search('test-file')
    await searchPage.waitForResults()
  })

  test('下载按钮初始状态', async () => {
    // Each result card should have a download button with "下载" text
    const firstButton = searchPage.resultArticles.first().getByRole('button', { name: '下载' })
    await expect(firstButton).toBeVisible()
    await expect(firstButton).toBeEnabled()
  })

  test('下载成功显示成功状态', async ({ page }) => {
    const firstArticle = searchPage.resultArticles.first()
    const downloadButton = firstArticle.getByRole('button')

    // Click download
    await downloadButton.click()

    // Should transition to success "成功"
    await expect(downloadButton).toContainText('成功', { timeout: 5_000 })

    // window.open should have been called
    const urls = await page.evaluate(() => (window as any).__openCalls as string[])
    expect(urls.length).toBeGreaterThanOrEqual(1)
    expect(urls[0]).toContain('example.com')

    // Should reset back to "下载" after 1.5s
    await expect(downloadButton).toContainText('下载', { timeout: 3_000 })
  })

  test('下载失败显示重试状态', async ({ page }) => {
    // Override mock for this test - return error
    await page.route('**/api/v1/app/download-url**', (route) =>
      route.fulfill({
        status: 502,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Bad Gateway' }),
      }),
    )

    const firstArticle = searchPage.resultArticles.first()
    const downloadButton = firstArticle.getByRole('button')

    await downloadButton.click()
    // Should show error state "重试"
    await expect(downloadButton).toContainText('重试', { timeout: 5_000 })
    // Button should still be clickable (enabled)
    await expect(downloadButton).toBeEnabled()
  })

  test('多个文件可同时下载', async ({ page }) => {
    const apiCalls: string[] = []
    await page.route('**/api/v1/app/download-url**', async (route) => {
      apiCalls.push(route.request().url())
      // Add small delay to keep buttons in loading state
      await new Promise((r) => setTimeout(r, 200))
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          file_id: 1001,
          download_url: 'https://example.com/fake-download.pdf',
        }),
      })
    })

    const first = searchPage.resultArticles.nth(0).getByRole('button')
    const second = searchPage.resultArticles.nth(1).getByRole('button')

    // Click both quickly
    await first.click()
    await second.click()

    // Both should show loading or success
    // Wait for both to complete
    await expect(first).toContainText('成功', { timeout: 5_000 })
    await expect(second).toContainText('成功', { timeout: 5_000 })

    // Two API calls should have been made
    expect(apiCalls.length).toBe(2)
  })
})

test.describe('边界场景', () => {
  let searchPage: SearchPage

  test.beforeAll(async () => {
    await seedMeilisearch()
  })

  test.afterAll(async () => {
    await clearMeilisearch()
  })

  test.beforeEach(async ({ page }) => {
    searchPage = new SearchPage(page)
    await searchPage.goto()
  })

  test('搜索特殊字符', async ({ page }) => {
    // Monitor for JS errors
    const errors: string[] = []
    page.on('pageerror', (err) => errors.push(err.message))

    const responsePromise = page.waitForResponse(
      (r) => r.url().includes('/api/v1/app/search') && r.status() === 200,
    )
    await searchPage.searchInput.fill('C++ & .NET')
    await searchPage.searchButton.click()
    const response = await responsePromise

    // Should have proper URL encoding
    expect(response.url()).toContain('query=')
    // No JS errors
    expect(errors).toHaveLength(0)
  })

  test('非常长的搜索查询', async ({ page }) => {
    const longQuery = 'a'.repeat(200)

    const responsePromise = page.waitForResponse(
      (r) => r.url().includes('/api/v1/app/search') && r.status() === 200,
    )
    await searchPage.searchInput.fill(longQuery)
    await searchPage.searchButton.click()
    const response = await responsePromise

    // Request should contain the full query
    expect(response.url()).toContain('query=')
    expect(response.status()).toBe(200)
  })

  test('快速连续搜索（防抖竞态）', async ({ page }) => {
    // Set up response listener BEFORE typing
    const responsePromise = page.waitForResponse(
      (r) => r.url().includes('/api/v1/app/search') && r.url().includes('query=abc') && r.status() === 200,
    )

    // Type rapidly
    await searchPage.searchInput.fill('a')
    await searchPage.searchInput.fill('ab')
    await searchPage.searchInput.fill('abc')

    // Wait for final search to complete
    const response = await responsePromise
    expect(response.status()).toBe(200)

    // The search input should show the final value
    await expect(searchPage.searchInput).toHaveValue('abc')
  })

  test('网络错误时显示错误状态', async ({ page }) => {
    // Mock search API to abort (network error)
    await page.route('**/api/v1/app/search**', (route) => route.abort())

    await searchPage.searchInput.fill('test')
    await searchPage.searchButton.click()

    // Should show error state
    await expect(page.locator('.border-rose-200')).toBeVisible({ timeout: 10_000 })
  })

  test('搜索框纯空格不触发搜索', async ({ page }) => {
    // Track API calls
    let apiCallCount = 0
    await page.route('**/api/v1/app/search**', (route) => {
      apiCallCount++
      return route.continue()
    })

    // Type only whitespace
    await searchPage.searchInput.fill('   ')

    // Wait longer than debounce (280ms)
    await page.waitForTimeout(500)

    // No API call should have been made
    expect(apiCallCount).toBe(0)

    // Should still be in initial state
    await expect(searchPage.heroMode).toBeVisible()
  })

  test('浏览器后退/前进导航', async ({ page }) => {
    // Search first
    await searchPage.search('test')
    await searchPage.waitForResults()

    // Navigate to admin
    await page.goto('/admin/')
    await expect(page).toHaveURL(/\/admin/)

    // Go back
    await page.goBack()
    await expect(page).toHaveURL(/\/$/)
  })
})
