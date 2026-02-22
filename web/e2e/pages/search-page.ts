import type { Locator, Page, Response } from '@playwright/test'

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

  async search(query: string): Promise<Response> {
    const responsePromise = this.page.waitForResponse(
      (r) => r.url().includes('/api/v1/app/search') && r.status() === 200,
    )
    await this.searchInput.fill(query)
    return responsePromise
  }

  async searchImmediate(query: string): Promise<Response> {
    const responsePromise = this.page.waitForResponse(
      (r) => r.url().includes('/api/v1/app/search') && r.status() === 200,
    )
    await this.searchInput.fill(query)
    await this.searchButton.click()
    return responsePromise
  }

  async waitForResults(): Promise<void> {
    await this.resultArticles.first().waitFor({ state: 'visible' })
  }

  async getResultCount(): Promise<number> {
    return this.resultArticles.count()
  }

  async scrollToLoadMore(): Promise<void> {
    await this.sentinel.scrollIntoViewIfNeeded()
  }
}
