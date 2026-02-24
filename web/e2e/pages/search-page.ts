import type { Locator, Page, Request, Response } from '@playwright/test'

type SearchRequestMatch = {
  query?: string
  page?: number
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function getRequestBody(request: Request): Record<string, unknown> | null {
  try {
    const body: unknown = request.postDataJSON()
    return isRecord(body) ? body : null
  } catch {
    return null
  }
}

function isConnectSearchResponse(
  response: Response,
  match?: SearchRequestMatch,
): boolean {
  if (
    !response.url().includes('/npan.v1.AppService/AppSearch') ||
    response.request().method() !== 'POST' ||
    response.status() !== 200
  ) {
    return false
  }
  if (!match) {
    return true
  }

  const body = getRequestBody(response.request())
  if (!body) {
    return false
  }

  if (match.query !== undefined && body.query !== match.query) {
    return false
  }
  if (match.page !== undefined && String(body.page) !== String(match.page)) {
    return false
  }

  return true
}

export class SearchPage {
  readonly page: Page
  readonly searchInput: Locator
  readonly searchButton: Locator
  readonly clearButton: Locator
  readonly resultArticles: Locator
  readonly sentinel: Locator
  readonly statusText: Locator
  readonly heroMode: Locator
  readonly dockedMode: Locator

  constructor(page: Page) {
    this.page = page
    this.searchInput = page.getByRole('searchbox')
    this.searchButton = page.getByRole('button', { name: '搜索', exact: true })
    this.clearButton = page.getByRole('button', { name: '清空搜索' })
    this.resultArticles = page.locator('article')
    this.sentinel = page.locator('.h-2').last()
    this.statusText = page.locator('.search-card p').last()
    this.heroMode = page.locator('.mode-hero')
    this.dockedMode = page.locator('.mode-docked')
  }

  async goto(): Promise<void> {
    await this.page.goto('/')
  }

  waitForSearchResponse(match?: SearchRequestMatch, timeout = 5_000): Promise<Response> {
    return this.page.waitForResponse((r) => isConnectSearchResponse(r, match), {
      timeout,
    })
  }

  async search(query: string): Promise<Response> {
    const responsePromise = this.waitForSearchResponse({ query })
    await this.searchInput.fill(query)
    return responsePromise
  }

  async searchImmediate(query: string): Promise<Response> {
    const responsePromise = this.waitForSearchResponse({ query })
    await this.searchInput.fill(query)
    await this.searchButton.click()
    return responsePromise
  }

  async waitForResults(): Promise<void> {
    await this.resultArticles.first().waitFor({ state: 'visible', timeout: 5_000 })
  }

  async getResultCount(): Promise<number> {
    return this.resultArticles.count()
  }

  async scrollToLoadMore(): Promise<void> {
    await this.sentinel.scrollIntoViewIfNeeded()
  }
}
