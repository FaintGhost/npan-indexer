import type { Locator, Page, Request, Response } from '@playwright/test'

type SearchRequestMatch = {
  query?: string
  page?: number
}

type PublicSearchRequestMatch = {
  query?: string
  filterContains?: string
}

type DownloadRequestMatch = {
  fileId?: string | number
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

function isPublicSearchResponse(
  response: Response,
  match?: PublicSearchRequestMatch,
): boolean {
  const isMeili = response.url().includes('/multi-search')
  const isTypesense = response.url().includes('/documents/search')
  if (!isMeili && !isTypesense) {
    return false
  }
  if (response.status() !== 200) {
    return false
  }
  if (isMeili && response.request().method() !== 'POST') {
    return false
  }
  if (isTypesense && response.request().method() !== 'GET') {
    return false
  }
  if (!match) {
    return true
  }

  if (isMeili) {
    const body = getRequestBody(response.request())
    if (!body || !Array.isArray(body.queries) || !isRecord(body.queries[0])) {
      return false
    }

    const firstQuery = body.queries[0]
    if (match.query !== undefined && firstQuery.q !== match.query) {
      return false
    }
    if (
      match.filterContains !== undefined
      && typeof firstQuery.filter === 'string'
      && !firstQuery.filter.includes(match.filterContains)
    ) {
      return false
    }
    if (
      match.filterContains !== undefined
      && Array.isArray(firstQuery.filter)
      && !JSON.stringify(firstQuery.filter).includes(match.filterContains)
    ) {
      return false
    }
    if (
      match.filterContains !== undefined
      && typeof firstQuery.filter !== 'string'
      && !Array.isArray(firstQuery.filter)
    ) {
      return false
    }
    return true
  }

  const url = new URL(response.url())
  if (match.query !== undefined && url.searchParams.get('q') !== match.query) {
    return false
  }
  if (
    match.filterContains !== undefined
    && !url.searchParams.get('filter_by')?.includes(match.filterContains)
  ) {
    return false
  }

  return true
}

function isDownloadResponse(
  response: Response,
  match?: DownloadRequestMatch,
): boolean {
  if (
    !response.url().includes('/npan.v1.AppService/AppDownloadURL') ||
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

  if (match.fileId !== undefined && String(body.fileId) !== String(match.fileId)) {
    return false
  }

  return true
}

export class SearchPage {
  readonly page: Page
  readonly searchHeader: Locator
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
    this.searchHeader = page.locator('header.search-stage')
    this.searchInput = page.getByRole('searchbox')
    this.searchButton = page.getByRole('button', { name: '搜索', exact: true })
    this.clearButton = page.getByRole('button', { name: '清空搜索' })
    this.resultArticles = page.locator('article')
    this.sentinel = page.locator('.h-2').last()
    this.statusText = page.locator('header.search-stage .mt-3 p.text-xs').first()
    this.heroMode = page.locator('.mode-hero')
    this.dockedMode = page.locator('.mode-docked')
  }

  async goto(path = '/'): Promise<void> {
    let lastError: unknown
    for (let attempt = 0; attempt < 2; attempt += 1) {
      try {
        await this.page.goto(path, { waitUntil: 'domcontentloaded', timeout: 12_000 })
        await this.searchInput.waitFor({ state: 'visible', timeout: 5_000 })
        return
      } catch (error) {
        lastError = error
        if (attempt === 1) {
          break
        }
        await this.page.waitForTimeout(400)
      }
    }
    throw lastError
  }

  waitForSearchResponse(match?: SearchRequestMatch, timeout = 5_000): Promise<Response> {
    return this.page.waitForResponse((r) => isConnectSearchResponse(r, match), {
      timeout,
    })
  }

  waitForPublicSearchResponse(
    match?: PublicSearchRequestMatch,
    timeout = 5_000,
  ): Promise<Response> {
    return this.page.waitForResponse((r) => isPublicSearchResponse(r, match), {
      timeout,
    })
  }

  waitForDownloadResponse(
    match?: DownloadRequestMatch,
    timeout = 5_000,
  ): Promise<Response> {
    return this.page.waitForResponse((r) => isDownloadResponse(r, match), {
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

  async searchPublicImmediate(query: string): Promise<Response> {
    const responsePromise = this.waitForPublicSearchResponse({ query })
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
