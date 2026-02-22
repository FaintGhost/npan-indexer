import { test, expect, ADMIN_API_KEY } from '../fixtures/auth'
import { AdminPage } from '../pages/admin-page'

test.describe('Admin 认证流程', () => {
  let adminPage: AdminPage

  test.beforeEach(async ({ page }) => {
    adminPage = new AdminPage(page)
  })

  // Test 1: 未认证时显示 API Key 对话框
  test('未认证时显示 API Key 对话框', async ({ page }) => {
    await adminPage.goto()
    // Dialog should be visible
    await expect(adminPage.dialog).toBeVisible()
    // Password input should be visible
    await expect(adminPage.apiKeyInput).toBeVisible()
    // Submit button should show "确认"
    await expect(adminPage.submitButton).toBeVisible()
    await expect(adminPage.submitButton).toHaveText('确认')
  })

  // Test 2: 空 API Key 显示本地错误
  test('空 API Key 显示本地错误', async ({ page }) => {
    await adminPage.goto()
    // Click submit without entering key
    await adminPage.submitButton.click()
    // Should show local error
    await expect(page.getByText('请输入 API Key')).toBeVisible()
    // Dialog should still be visible
    await expect(adminPage.dialog).toBeVisible()
  })

  // Test 3: 错误 API Key 显示服务端错误
  test('错误 API Key 显示服务端错误', async ({ page }) => {
    await adminPage.goto()
    // Enter wrong key and submit
    await adminPage.submitApiKey('wrong-key-00000')
    // Wait for server response and error
    await expect(page.getByText('API Key 无效')).toBeVisible({ timeout: 10_000 })
    // Dialog should still be visible
    await expect(adminPage.dialog).toBeVisible()
  })

  // Test 4: 正确 API Key 进入管理界面
  test('正确 API Key 进入管理界面', async ({ page }) => {
    await adminPage.goto()
    // Enter correct key
    await adminPage.submitApiKey(ADMIN_API_KEY)
    // Dialog should disappear
    await adminPage.waitForAuthComplete()
    // Should show sync management UI
    await expect(page.getByRole('heading', { name: '同步管理' })).toBeVisible()
    // Check localStorage
    const storedKey = await page.evaluate(() => localStorage.getItem('npan_admin_api_key'))
    expect(storedKey).toBe(ADMIN_API_KEY)
  })

  // Test 5: 刷新页面保持认证状态
  test('刷新页面保持认证状态', async ({ authenticatedPage }) => {
    const authAdminPage = new AdminPage(authenticatedPage)
    await authAdminPage.goto()
    // Should NOT show dialog
    await expect(authAdminPage.dialog).not.toBeVisible()
    // Should show sync management
    await expect(authenticatedPage.getByRole('heading', { name: '同步管理' })).toBeVisible()
    // Reload page
    await authenticatedPage.reload()
    // Should still be authenticated
    await expect(authAdminPage.dialog).not.toBeVisible()
    await expect(authenticatedPage.getByRole('heading', { name: '同步管理' })).toBeVisible()
  })

  // Test 6: 返回搜索链接
  test('返回搜索链接导航到首页', async ({ authenticatedPage }) => {
    const authAdminPage = new AdminPage(authenticatedPage)
    await authAdminPage.goto()
    await expect(authAdminPage.dialog).not.toBeVisible()
    // Click back to search
    await authAdminPage.backToSearchLink.click()
    // Should navigate to /
    await expect(authenticatedPage).toHaveURL(/\/$/)
  })
})

test.describe('Admin 同步控制', () => {
  let adminPage: AdminPage

  test.beforeEach(async ({ authenticatedPage }) => {
    adminPage = new AdminPage(authenticatedPage)
    await adminPage.goto()
    // Wait for auth to be complete (dialog should not show)
    await expect(adminPage.dialog).not.toBeVisible()
    await expect(authenticatedPage.getByRole('heading', { name: '同步管理' })).toBeVisible()
  })

  test('显示同步模式选择器', async ({ authenticatedPage }) => {
    // All three mode buttons should be visible
    await expect(adminPage.modeButtons.auto).toBeVisible()
    await expect(adminPage.modeButtons.full).toBeVisible()
    await expect(adminPage.modeButtons.incremental).toBeVisible()
    // "自适应" should be selected by default (has bg-white class)
    await expect(adminPage.modeButtons.auto).toHaveClass(/bg-white/)
  })

  test('选择全量模式', async ({ authenticatedPage }) => {
    await adminPage.selectMode('full')
    // "全量" should be selected
    await expect(adminPage.modeButtons.full).toHaveClass(/bg-white/)
    // Others should not be selected
    await expect(adminPage.modeButtons.auto).not.toHaveClass(/bg-white/)
    await expect(adminPage.modeButtons.incremental).not.toHaveClass(/bg-white/)
  })

  test('启动同步发送正确请求', async ({ authenticatedPage }) => {
    // Select full mode
    await adminPage.selectMode('full')

    // Monitor the POST request
    const syncRequest = authenticatedPage.waitForRequest(
      (req) => req.url().includes('/api/v1/admin/sync') && req.method() === 'POST',
    )

    // Click start sync
    await adminPage.startSyncButton.click()
    const request = await syncRequest

    // Verify request has correct headers and body
    expect(request.headers()['x-api-key']).toBeTruthy()
    const body = request.postDataJSON()
    expect(body.mode).toBe('full')
  })

  test('取消同步触发确认对话框', async ({ authenticatedPage }) => {
    // Start sync first
    const syncResponse = authenticatedPage.waitForResponse(
      (r) => r.url().includes('/api/v1/admin/sync') && r.request().method() === 'POST',
    )
    await adminPage.startSyncButton.click()
    await syncResponse

    // Wait for cancel button to appear (if sync is running)
    // Note: with dummy NPA_TOKEN, sync may fail fast. Only test cancel if button appears
    try {
      await adminPage.cancelSyncButton.waitFor({ state: 'visible', timeout: 3_000 })
    } catch {
      // If cancel button doesn't appear (sync already finished), skip this test
      test.skip()
      return
    }

    // Set up dialog handler - accept the confirmation
    authenticatedPage.once('dialog', async (dialog) => {
      expect(dialog.message()).toContain('确认取消同步')
      await dialog.accept()
    })

    // Monitor DELETE request
    const deleteRequest = authenticatedPage.waitForRequest(
      (req) => req.url().includes('/api/v1/admin/sync') && req.method() === 'DELETE',
    )

    // Click cancel
    await adminPage.cancelSyncButton.click()
    await deleteRequest
  })

  test('取消确认框点击取消不发请求', async ({ authenticatedPage }) => {
    // Start sync first
    const syncResponse = authenticatedPage.waitForResponse(
      (r) => r.url().includes('/api/v1/admin/sync') && r.request().method() === 'POST',
    )
    await adminPage.startSyncButton.click()
    await syncResponse

    // Wait for cancel button
    try {
      await adminPage.cancelSyncButton.waitFor({ state: 'visible', timeout: 3_000 })
    } catch {
      test.skip()
      return
    }

    // Set up dialog handler - dismiss (cancel)
    authenticatedPage.once('dialog', async (dialog) => {
      await dialog.dismiss()
    })

    // Track if DELETE is sent
    let deleteSent = false
    authenticatedPage.on('request', (req) => {
      if (req.url().includes('/api/v1/admin/sync') && req.method() === 'DELETE') {
        deleteSent = true
      }
    })

    // Click cancel button
    await adminPage.cancelSyncButton.click()

    // Wait briefly and verify no DELETE was sent
    await authenticatedPage.waitForTimeout(1000)
    expect(deleteSent).toBe(false)
  })
})

test.describe('Admin 边界场景', () => {
  test('认证过期后显示对话框', async ({ page }) => {
    const adminPage = new AdminPage(page)

    // Navigate to admin first, then set localStorage
    await page.goto('/admin/')
    await page.evaluate((key: string) => {
      localStorage.setItem('npan_admin_api_key', key)
    }, ADMIN_API_KEY)

    // Reload to pick up the key
    await page.reload()
    await expect(adminPage.dialog).not.toBeVisible()
    await expect(page.getByRole('heading', { name: '同步管理' })).toBeVisible()

    // Clear localStorage to simulate expiry
    await page.evaluate(() => localStorage.removeItem('npan_admin_api_key'))

    // Reload page
    await page.reload()

    // Should show API Key dialog again
    await expect(adminPage.dialog).toBeVisible()
  })
})
