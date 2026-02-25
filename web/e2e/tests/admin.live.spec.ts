import { test, expect } from '../fixtures/auth'
import { AdminPage } from '../pages/admin-page'

const isLive = process.env.E2E_LIVE === '1'

function requireLive(testName: string) {
  test.skip(!isLive, `${testName}: 需要设置 E2E_LIVE=1 并使用真实 NPA_TOKEN 环境`)
}

test.describe('Admin Live（真实数据）', () => {
  test.describe.configure({ mode: 'serial' })

  test('live: 全量启动可进入运行态', async ({ authenticatedPage }) => {
    requireLive('live: 全量启动可进入运行态')

    const adminPage = new AdminPage(authenticatedPage)
    await adminPage.goto()
    await expect(adminPage.dialog).not.toBeVisible()
    await expect(authenticatedPage.getByRole('heading', { name: '同步管理' })).toBeVisible()

    const startReq = authenticatedPage.waitForRequest(
      (req) =>
        req.method() === 'POST' && req.url().includes('/npan.v1.AdminService/StartSync'),
      { timeout: 15_000 },
    )

    await adminPage.startSyncButton.click()
    await startReq

    await expect(adminPage.startSyncButton).toBeDisabled({ timeout: 10_000 })

    const cancelVisible = await adminPage.cancelSyncButton
      .waitFor({ state: 'visible', timeout: 10_000 })
      .then(() => true)
      .catch(() => false)

    if (cancelVisible) {
      const cancelReq = authenticatedPage.waitForRequest(
        (req) =>
          req.method() === 'POST' && req.url().includes('/npan.v1.AdminService/CancelSync'),
        { timeout: 15_000 },
      )
      await adminPage.cancelSyncButton.click()
      await authenticatedPage.getByRole('button', { name: '确认取消' }).click()
      await cancelReq
      await expect(authenticatedPage.getByText(/已发送取消请求|同步取消信号已发送/)).toBeVisible({
        timeout: 10_000,
      })
    }
  })

  test('live: InspectRoots 真实链路可返回并展示结果', async ({ authenticatedPage }) => {
    requireLive('live: InspectRoots 真实链路可返回并展示结果')
    test.setTimeout(180_000)

    const adminPage = new AdminPage(authenticatedPage)
    await adminPage.goto()
    await expect(adminPage.dialog).not.toBeVisible()
    await expect(authenticatedPage.getByRole('heading', { name: '同步管理' })).toBeVisible()

    const rootDetailsHeader = authenticatedPage.getByRole('button', { name: /根目录详情 \(/ })
    const hasRootCatalog = await rootDetailsHeader
      .waitFor({ state: 'visible', timeout: 5_000 })
      .then(() => true)
      .catch(() => false)

    if (!hasRootCatalog) {
      test.skip(true, 'live 环境尚未建立根目录 catalog（请先跑一次全量以生成根目录详情）')
      return
    }

    const inspectButton = authenticatedPage.getByRole('button', { name: /刷新目录详情/ })
    await expect(inspectButton).toBeEnabled()

    const startedAt = Date.now()
    const inspectResp = authenticatedPage.waitForResponse(
      (resp) =>
        resp.request().method() === 'POST' &&
        resp.url().includes('/npan.v1.AdminService/InspectRoots'),
      { timeout: 120_000 },
    )

    await inspectButton.click()
    const response = await inspectResp
    const finishedAt = Date.now()

    expect(response.status()).toBe(200)

    const body = (await response.json()) as { items?: unknown[]; errors?: unknown[] }
    const items = Array.isArray(body.items) ? body.items.length : 0
    const errors = Array.isArray(body.errors) ? body.errors.length : 0

    await expect(authenticatedPage.getByText(/目录详情已拉取：成功/)).toBeVisible({
      timeout: 120_000,
    })

    if ((await rootDetailsHeader.locator('span', { hasText: '展开' }).count()) > 0) {
      await rootDetailsHeader.click()
    }

    await expect(authenticatedPage.getByText(/^估计\s/).first()).toBeVisible({ timeout: 30_000 })

    // 输出真实耗时用于人工观察（不做硬阈值，避免被上游波动误伤）
    console.log(
      `[live-admin] InspectRoots latency=${finishedAt - startedAt}ms, items=${items}, errors=${errors}`,
    )
  })
})
