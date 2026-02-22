import { test as base, expect } from '@playwright/test'
import type { Page } from '@playwright/test'

export const ADMIN_API_KEY = process.env.E2E_ADMIN_API_KEY ?? 'ci-test-admin-api-key-1234'

export const test = base.extend<{ authenticatedPage: Page }>({
  authenticatedPage: async ({ context }, use) => {
    const apiKey = ADMIN_API_KEY
    await context.addInitScript((key: string) => {
      localStorage.setItem('npan_admin_api_key', key)
    }, apiKey)

    const page = await context.newPage()
    await page.goto('about:blank')
    await use(page)
  },
})

export { expect }
