import type { Page } from '@playwright/test'
import { test, expect } from '../fixtures/auth'
import { SearchPage } from '../pages/search-page'
import { seedSearchBackend, clearSearchBackend } from '../fixtures/seed'

const PUBLIC_MEILI_HOST = process.env.MEILI_PUBLIC_SEARCH_HOST ?? process.env.MEILI_HOST ?? 'http://localhost:7700'
const PUBLIC_MEILI_INDEX = process.env.MEILI_PUBLIC_SEARCH_INDEX ?? process.env.MEILI_INDEX ?? 'npan_items'
const PUBLIC_MEILI_SEARCH_API_KEY = process.env.MEILI_PUBLIC_SEARCH_API_KEY ?? 'ci-test-public-search-key-5678'

interface PublicSearchDoc {
  source_id: number
  name: string
  file_category: 'doc' | 'image' | 'video' | 'archive' | 'other'
}

const PUBLIC_SEARCH_DOCS: PublicSearchDoc[] = [
  {
    source_id: 9001,
    name: 'quarterly-report-2024.pdf',
    file_category: 'doc',
  },
  {
    source_id: 9002,
    name: 'project-design-spec.docx',
    file_category: 'doc',
  },
  {
    source_id: 9003,
    name: 'architecture-diagram.png',
    file_category: 'image',
  },
]

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function getStringField(record: Record<string, unknown>, key: string): string | null {
  const value = record[key]
  if (typeof value === 'string') {
    return value
  }
  if (typeof value === 'number') {
    return String(value)
  }
  return null
}

async function enablePublicSearchBootstrap(page: Page): Promise<void> {
  await page.route('**/npan.v1.AppService/GetSearchConfig**', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        host: PUBLIC_MEILI_HOST,
        indexName: PUBLIC_MEILI_INDEX,
        searchApiKey: PUBLIC_MEILI_SEARCH_API_KEY,
        instantsearchEnabled: true,
      }),
    }),
  )

  await page.route('**/multi-search**', async (route) => {
    const payload: unknown = route.request().postDataJSON()
    const postData = isRecord(payload) ? payload : {}
    const queries = Array.isArray(postData.queries) ? postData.queries : []
    const queryPayload = isRecord(queries[0]) ? queries[0] : {}
    const rawQuery = typeof queryPayload.q === 'string' ? queryPayload.q : ''
    const query = rawQuery.trim().toLowerCase()

    const rawFilter = queryPayload.filter
    const filterText = typeof rawFilter === 'string'
      ? rawFilter
      : Array.isArray(rawFilter)
        ? JSON.stringify(rawFilter)
        : ''

    const matchedDocs = PUBLIC_SEARCH_DOCS.filter((doc) => doc.name.toLowerCase().includes(query))
    const docs = filterText.includes('file_category')
      ? matchedDocs.filter((doc) => filterText.includes(doc.file_category))
      : matchedDocs

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        results: [
          {
            indexUid: PUBLIC_MEILI_INDEX,
            hits: docs.map((doc) => ({
              id: String(doc.source_id),
              doc_id: `file_${doc.source_id}`,
              source_id: doc.source_id,
              type: 'file',
              name: doc.name,
              path_text: `/${doc.name}`,
              parent_id: 0,
              modified_at: 1700000000,
              created_at: 1700000000,
              size: 1024,
              sha1: `sha1-${doc.source_id}`,
              in_trash: false,
              is_deleted: false,
              file_category: doc.file_category,
              downloadUrl: 'https://example.com/hit-download.pdf',
            })),
            query: rawQuery,
            processingTimeMs: 1,
            limit: 20,
            offset: 0,
            estimatedTotalHits: docs.length,
            facetDistribution: {
              file_category: docs.reduce<Record<string, number>>((acc, doc) => {
                acc[doc.file_category] = (acc[doc.file_category] ?? 0) + 1
                return acc
              }, {}),
            },
          },
        ],
      }),
    })
  })
}

test.describe('搜索流程', () => {
  let searchPage: SearchPage

  test.beforeAll(async () => {
    await seedSearchBackend()
  })

  test.afterAll(async () => {
    await clearSearchBackend()
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
    const responsePromise = searchPage.waitForSearchResponse({ query: 'design' })
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
    // Search for "test" should match 35 bulk docs
    await searchPage.search('test')
    await searchPage.waitForResults()
    // First page should load 30 results
    await expect(searchPage.resultArticles).toHaveCount(30, { timeout: 10_000 })
    // Scroll to sentinel to trigger infinite scroll
    const secondPageResponse = searchPage.waitForSearchResponse({
      query: 'test',
      page: 2,
    })
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

  test('docked 吸顶头部使用不透明背景并遮挡结果内容', async () => {
    await searchPage.search('test')
    await searchPage.waitForResults()
    await expect(searchPage.dockedMode).toBeVisible()
    await expect(searchPage.searchHeader).toHaveClass(/search-stage-opaque/)

    const backgroundColor = await searchPage.searchHeader.evaluate((element) =>
      window.getComputedStyle(element).backgroundColor,
    )
    expect(backgroundColor).not.toBe('rgba(0, 0, 0, 0)')

    const topResult = searchPage.resultArticles.first()
    const [headerBox, resultBox] = await Promise.all([
      searchPage.searchHeader.boundingBox(),
      topResult.boundingBox(),
    ])

    expect(headerBox).not.toBeNull()
    expect(resultBox).not.toBeNull()
    if (!headerBox || !resultBox) {
      return
    }

    expect(resultBox.y).toBeGreaterThanOrEqual(headerBox.y + headerBox.height - 1)
  })

  // Test 10: 分类筛选与 URL 参数同步
  test('分类筛选与 URL 参数同步', async ({ page }) => {
    await searchPage.search('test')
    await searchPage.waitForResults()
    await expect(searchPage.resultArticles).toHaveCount(30, { timeout: 10_000 })

    const imageFilter = page.getByRole('radio', { name: '图片' })
    const allFilter = page.getByRole('radio', { name: '全部' })

    await imageFilter.click()
    await expect(page).toHaveURL(/file_category=image/)
    await expect(page.getByRole('heading', { name: '未找到相关文件' })).toBeVisible()
    await expect(searchPage.resultArticles).toHaveCount(0)

    await allFilter.click()
    await expect(page).not.toHaveURL(/file_category=/)
    await expect(searchPage.resultArticles).toHaveCount(35, { timeout: 10_000 })
  })

  test('URL 可恢复 public query 与 file_category refinement', async ({ page }) => {
    await enablePublicSearchBootstrap(page)
    searchPage = new SearchPage(page)

    const responsePromise = searchPage.waitForPublicSearchResponse({
      query: 't',
      filterContains: 'file_category',
    }, 10_000)
    await searchPage.goto('/?query=t&file_category=doc')
    const response = await responsePromise

    expect(response.status()).toBe(200)
    await expect(searchPage.searchInput).toHaveValue('t')
    await expect(page.getByRole('radio', { name: '文档' })).toBeChecked()
    await expect(page.getByTitle('quarterly-report-2024.pdf')).toBeVisible()
    await expect(page.getByTitle('project-design-spec.docx')).toBeVisible()
    await expect(page.getByTitle('architecture-diagram.png')).toHaveCount(0)
  })

  test('浏览器后退/前进可恢复 public search 视图', async ({ page }) => {
    await enablePublicSearchBootstrap(page)
    searchPage = new SearchPage(page)
    await searchPage.goto()

    const initialSearch = await searchPage.searchPublicImmediate('t')
    expect(initialSearch.status()).toBe(200)
    await expect(searchPage.searchInput).toHaveValue('t')
    await expect(page.getByTitle('quarterly-report-2024.pdf')).toBeVisible()

    const imageFilterResponse = searchPage.waitForPublicSearchResponse({
      query: 't',
      filterContains: 'file_category',
    }, 10_000)
    await page.getByRole('radio', { name: '图片' }).click()
    await imageFilterResponse

    await expect(page).toHaveURL(/file_category=image/)
    await expect(page.getByRole('radio', { name: '图片' })).toBeChecked()
    await expect(page.getByTitle('architecture-diagram.png')).toBeVisible()
    await expect(page.getByTitle('quarterly-report-2024.pdf')).toHaveCount(0)

    await page.goBack()

    await expect(page).not.toHaveURL(/file_category=image/)
    await expect(searchPage.searchInput).toHaveValue('t')
    await expect(page.getByRole('radio', { name: '全部' })).toBeChecked()
    await expect(page.getByTitle('quarterly-report-2024.pdf')).toBeVisible()

    await page.goForward()

    await expect(page).toHaveURL(/file_category=image/)
    await expect(page.getByRole('radio', { name: '图片' })).toBeChecked()
    await expect(page.getByTitle('architecture-diagram.png')).toBeVisible()
    await expect(page.getByTitle('quarterly-report-2024.pdf')).toHaveCount(0)
  })
})

test.describe('下载流程', () => {
  let searchPage: SearchPage

  test.beforeAll(async () => {
    await seedSearchBackend()
  })

  test.afterAll(async () => {
    await clearSearchBackend()
  })

  test.beforeEach(async ({ page }) => {
    searchPage = new SearchPage(page)

    // Intercept window.open
    await page.addInitScript(() => {
      const calls: string[] = []
      Reflect.set(window, '__openCalls', calls)
      window.open = (url?: string | URL) => {
        if (url) calls.push(String(url))
        return null
      }
    })

    // Mock download API - success by default
    await page.route('**/npan.v1.AppService/AppDownloadURL**', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          result: {
            fileId: '1001',
            downloadUrl: 'https://example.com/fake-download.pdf',
          },
        }),
      }),
    )

    await searchPage.goto()
    // Search and get results first
    await searchPage.search('test')
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
    const urls = await page.evaluate(() => {
      const openCalls = Reflect.get(window, '__openCalls')
      return Array.isArray(openCalls)
        ? openCalls.filter((value): value is string => typeof value === 'string')
        : []
    })
    expect(urls.length).toBeGreaterThanOrEqual(1)
    expect(urls[0]).toContain('example.com')

    // Should reset back to "下载" after 1.5s
    await expect(downloadButton).toContainText('下载', { timeout: 3_000 })
  })

  test('下载失败显示重试状态', async ({ page }) => {
    // Override mock for this test - return error
    await page.route('**/npan.v1.AppService/AppDownloadURL**', (route) =>
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

  test('public InstantSearch 下载仍调用 AppDownloadURL 而不是依赖 hit 内地址', async ({ page }) => {
    await enablePublicSearchBootstrap(page)
    searchPage = new SearchPage(page)

    const downloadRequests: Array<Record<string, unknown>> = []
    await page.route('**/npan.v1.AppService/AppDownloadURL**', async (route) => {
      const payload: unknown = route.request().postDataJSON()
      if (isRecord(payload)) {
        downloadRequests.push(payload)
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          result: {
            fileId: '9001',
            downloadUrl: 'https://example.com/public-rpc-download.pdf',
          },
        }),
      })
    })

    await searchPage.goto()
    const publicSearchResponse = searchPage.waitForPublicSearchResponse({ query: 'quarterly' }, 10_000)
    await searchPage.searchInput.fill('quarterly')
    await searchPage.searchButton.click()
    await publicSearchResponse

    const resultCard = page.locator('article').filter({ has: page.getByTitle('quarterly-report-2024.pdf') }).first()
    await expect(resultCard).toBeVisible()

    const downloadResponse = searchPage.waitForDownloadResponse({ fileId: 9001 }, 5_000)
    await resultCard.getByRole('button', { name: '下载' }).click()
    await downloadResponse

    expect(downloadRequests).toEqual([{ fileId: '9001' }])

    const urls = await page.evaluate(() => {
      const openCalls = Reflect.get(window, '__openCalls')
      return Array.isArray(openCalls)
        ? openCalls.filter((value): value is string => typeof value === 'string')
        : []
    })
    expect(urls).toContain('https://example.com/public-rpc-download.pdf')
    expect(urls).not.toContain('https://example.com/hit-download.pdf')
  })

  test('多个文件可同时下载', async ({ page }) => {
    const apiCalls: string[] = []
    await page.route('**/npan.v1.AppService/AppDownloadURL**', async (route) => {
      apiCalls.push(route.request().url())
      // Add small delay to keep buttons in loading state
      await new Promise((r) => setTimeout(r, 200))
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          result: {
            fileId: '1001',
            downloadUrl: 'https://example.com/fake-download.pdf',
          },
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
    await seedSearchBackend()
  })

  test.afterAll(async () => {
    await clearSearchBackend()
  })

  test.beforeEach(async ({ page }) => {
    searchPage = new SearchPage(page)
    await searchPage.goto()
  })

  test('搜索特殊字符', async ({ page }) => {
    // Monitor for JS errors
    const errors: string[] = []
    page.on('pageerror', (err) => errors.push(err.message))

    const responsePromise = searchPage.waitForSearchResponse({ query: 'C++ & .NET' })
    await searchPage.searchInput.fill('C++ & .NET')
    await searchPage.searchButton.click()
    const response = await responsePromise

    const payload: unknown = response.request().postDataJSON()
    expect(isRecord(payload)).toBe(true)
    if (!isRecord(payload)) {
      return
    }
    expect(getStringField(payload, 'query')).toBe('C++ & .NET')

    // No JS errors
    expect(errors).toHaveLength(0)
  })

  test('非常长的搜索查询', async () => {
    const longQuery = 'a'.repeat(200)

    const responsePromise = searchPage.waitForSearchResponse({ query: longQuery })
    await searchPage.searchInput.fill(longQuery)
    await searchPage.searchButton.click()
    const response = await responsePromise

    const payload: unknown = response.request().postDataJSON()
    expect(isRecord(payload)).toBe(true)
    if (!isRecord(payload)) {
      return
    }
    expect(getStringField(payload, 'query')).toBe(longQuery)
    expect(response.status()).toBe(200)
  })

  test('快速连续搜索（防抖竞态）', async () => {
    // Set up response listener BEFORE typing
    const responsePromise = searchPage.waitForSearchResponse({ query: 'abc' })

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
    await page.route('**/npan.v1.AppService/AppSearch**', (route) => route.abort())

    await searchPage.searchInput.fill('test')
    await searchPage.searchButton.click()

    // Should show error state
    await expect(page.getByRole('heading', { name: '加载出错了' })).toBeVisible({ timeout: 5_000 })
  })

  test('搜索框纯空格不触发搜索', async ({ page }) => {
    // Track API calls
    let apiCallCount = 0
    await page.route('**/npan.v1.AppService/AppSearch**', (route) => {
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
