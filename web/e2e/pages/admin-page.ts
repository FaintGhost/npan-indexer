import type { Locator, Page } from '@playwright/test'

export class AdminPage {
  readonly page: Page
  readonly apiKeyInput: Locator
  readonly submitButton: Locator
  readonly startSyncButton: Locator
  readonly cancelSyncButton: Locator
  readonly modeButtons: {
    auto: Locator
    full: Locator
    incremental: Locator
  }
  readonly dialog: Locator
  readonly syncMessage: Locator
  readonly errorMessage: Locator
  readonly backToSearchLink: Locator
  readonly syncStatusBadge: Locator

  constructor(page: Page) {
    this.page = page
    this.apiKeyInput = page.locator('input[type="password"]')
    this.submitButton = page.locator('button[type="submit"]')
    this.startSyncButton = page.getByRole('button', { name: /启动同步|同步进行中|启动中/ })
    this.cancelSyncButton = page.getByRole('button', { name: '取消同步' })
    this.modeButtons = {
      auto: page.getByRole('button', { name: '自适应' }),
      full: page.getByRole('button', { name: '全量' }),
      incremental: page.getByRole('button', { name: '增量' }),
    }
    this.dialog = page.locator('.fixed.inset-0')
    this.syncMessage = page.locator('.border-emerald-200')
    this.errorMessage = page.locator('.border-rose-200')
    this.backToSearchLink = page.getByRole('link', { name: /返回搜索/ })
    this.syncStatusBadge = page.locator('text=运行中')
  }

  async goto(): Promise<void> {
    await this.page.goto('/admin/')
  }

  async submitApiKey(key: string): Promise<void> {
    await this.apiKeyInput.fill(key)
    await this.submitButton.click()
  }

  async injectApiKey(key: string): Promise<void> {
    await this.page.context().addInitScript((k: string) => {
      localStorage.setItem('npan_admin_api_key', k)
    }, key)
  }

  async selectMode(mode: 'auto' | 'full' | 'incremental'): Promise<void> {
    await this.modeButtons[mode].click()
  }

  async waitForAuthComplete(): Promise<void> {
    await this.dialog.waitFor({ state: 'hidden' })
  }
}
