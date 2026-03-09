import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { createTestProvider } from '../tests/test-providers'
import { AdminSyncPage } from './admin-sync-page'

const STORAGE_KEY = 'npan_admin_api_key'

function assertRecord(value: unknown): asserts value is Record<string, unknown> {
  if (typeof value !== 'object' || value === null) {
    throw new Error('expected payload to be an object')
  }
}

function getRecord(value: unknown): Record<string, unknown> {
  assertRecord(value)
  return value
}

function toProtoStatus(status: string) {
  switch (status) {
    case 'running':
      return 'SYNC_STATUS_RUNNING'
    case 'done':
      return 'SYNC_STATUS_DONE'
    case 'error':
      return 'SYNC_STATUS_ERROR'
    case 'cancelled':
      return 'SYNC_STATUS_CANCELLED'
    case 'interrupted':
      return 'SYNC_STATUS_INTERRUPTED'
    case 'idle':
    default:
      return 'SYNC_STATUS_IDLE'
  }
}

function toProtoMode(mode?: string) {
  switch (mode) {
    case 'full':
      return 'SYNC_MODE_FULL'
    case 'incremental':
      return 'SYNC_MODE_INCREMENTAL'
    default:
      return undefined
  }
}

function toConnectProgressResponse(progress: Record<string, unknown>) {
  return {
    state: {
      ...progress,
      status: toProtoStatus(String(progress.status ?? 'idle')),
      mode: toProtoMode(progress.mode ? String(progress.mode) : undefined),
    },
  }
}

async function advanceTimers(ms: number) {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(ms)
  })
}

const validProgress = {
  status: 'idle',
  mode: 'full',
  startedAt: 0,
  updatedAt: 0,
  roots: [],
  completedRoots: [],
  aggregateStats: {
    foldersVisited: 0,
    filesIndexed: 0,
    filesDiscovered: 0,
    skippedFiles: 0,
    pagesFetched: 0,
    failedRequests: 0,
    startedAt: 0,
    endedAt: 0,
  },
  rootProgress: {},
}

function mockAdminBasics(options?: {
  progress?: Record<string, unknown>
  indexDocumentCount?: number
  indexStatsStatus?: number
}) {
  const progress = options?.progress ?? validProgress
  const indexDocumentCount = options?.indexDocumentCount ?? 10
  const indexStatsStatus = options?.indexStatsStatus ?? 200

  server.use(
    http.post('/npan.v1.AdminService/GetSyncProgress', () => {
      return HttpResponse.json(toConnectProgressResponse(progress))
    }),
    http.post('/npan.v1.AdminService/GetIndexStats', () => {
      if (indexStatsStatus !== 200) {
        return HttpResponse.json({ code: 'internal', message: 'boom' }, { status: indexStatsStatus })
      }
      return HttpResponse.json({ documentCount: String(indexDocumentCount) })
    }),
    http.post('/npan.v1.AdminService/WatchSyncProgress', () => {
      return HttpResponse.json({ code: 'unimplemented' }, { status: 501 })
    }),
  )
}

describe('AdminSyncPage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows API key dialog when no stored key', () => {
    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })
    expect(screen.getByPlaceholderText(/API Key/i)).toBeInTheDocument()
  })

  it('shows admin panel when key is stored', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    mockAdminBasics()

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^启动同步$/ })).toBeInTheDocument()
    })
  })

  it('defaults to full mode and only renders two mode buttons', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    mockAdminBasics()

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: '全量' })).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: '增量' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '自适应' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: '全量' })).toHaveClass('bg-white')
  })

  it('shows incremental disabled hint while checking index status', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    server.use(
      http.post('/npan.v1.AdminService/GetSyncProgress', () => {
        return HttpResponse.json(toConnectProgressResponse(validProgress))
      }),
      http.post('/npan.v1.AdminService/GetIndexStats', async () => {
        await new Promise((resolve) => setTimeout(resolve, 200))
        return HttpResponse.json({ documentCount: '10' })
      }),
      http.post('/npan.v1.AdminService/WatchSyncProgress', () => {
        return HttpResponse.json({ code: 'unimplemented' }, { status: 501 })
      }),
    )

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    expect(await screen.findByText('正在检查索引状态...')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '增量' })).toBeDisabled()
  })

  it('shows empty-index hint and blocks incremental start', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    mockAdminBasics({ indexDocumentCount: 0 })

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await screen.findByText('请先执行一次全量索引')
    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: /^启动同步$/ }))
    expect(await screen.findByText('请先执行一次全量索引')).toBeInTheDocument()
  })

  it('shows unknown-index hint when index status request fails', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    mockAdminBasics({ indexStatsStatus: 500 })

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    expect(await screen.findByText('无法确认索引状态，请稍后重试')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '增量' })).toBeDisabled()
  })

  it('running state disables mode switch and start but keeps refresh and cancel', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')

    let inspectCalled = false
    mockAdminBasics({
      progress: {
        ...validProgress,
        status: 'running',
        roots: [1001],
        catalogRoots: [1001],
        catalogRootNames: { '1001': 'A' },
        catalogRootProgress: {
          '1001': {
            rootFolderId: 1001,
            status: 'running',
            estimatedTotalDocs: 11,
            stats: validProgress.aggregateStats,
            updatedAt: 0,
          },
        },
      },
    })
    server.use(
      http.post('/npan.v1.AdminService/InspectRoots', () => {
        inspectCalled = true
        return HttpResponse.json({ items: [], errors: [] })
      }),
    )

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await screen.findByText('运行中')
    expect(screen.getByRole('button', { name: '全量' })).toBeDisabled()
    expect(screen.getByRole('button', { name: '增量' })).toBeDisabled()
    expect(screen.getByRole('button', { name: /^启动同步$/ })).toBeDisabled()

    const user = userEvent.setup()
    const refreshBtn = screen.getByRole('button', { name: /刷新目录详情/i })
    expect(refreshBtn).toBeEnabled()
    await user.click(refreshBtn)
    await waitFor(() => expect(inspectCalled).toBe(true))
    expect(screen.getByRole('button', { name: '取消同步' })).toBeInTheDocument()
  })

  it('hides root selection switches in incremental mode', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    mockAdminBasics({
      progress: {
        ...validProgress,
        catalogRoots: [1001],
        catalogRootNames: { '1001': 'A' },
        catalogRootProgress: {
          '1001': {
            rootFolderId: 1001,
            status: 'done',
            estimatedTotalDocs: 11,
            stats: {
              ...validProgress.aggregateStats,
              filesIndexed: 10,
              foldersVisited: 1,
            },
            updatedAt: 0,
          },
        },
      },
      indexDocumentCount: 10,
    })

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await screen.findByRole('button', { name: '增量' })
    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: '增量' }))
    await user.click(screen.getByRole('button', { name: /展开/i }))
    expect(screen.queryByRole('switch', { name: /选择根目录/i })).not.toBeInTheDocument()
  })

  it('blocks force rebuild with scoped full selection', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    let startCalled = false
    mockAdminBasics({
      progress: {
        ...validProgress,
        catalogRoots: [1001, 1002],
        catalogRootNames: { '1001': 'A', '1002': 'B' },
        catalogRootProgress: {
          '1001': {
            rootFolderId: 1001,
            status: 'done',
            estimatedTotalDocs: 11,
            stats: { ...validProgress.aggregateStats, filesIndexed: 10, foldersVisited: 1 },
            updatedAt: 0,
          },
          '1002': {
            rootFolderId: 1002,
            status: 'done',
            estimatedTotalDocs: 21,
            stats: { ...validProgress.aggregateStats, filesIndexed: 20, foldersVisited: 1 },
            updatedAt: 0,
          },
        },
      },
    })
    server.use(
      http.post('/npan.v1.AdminService/StartSync', () => {
        startCalled = true
        return HttpResponse.json({ message: 'Sync started' })
      }),
    )

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await screen.findByText(/当前已勾选/) 
    const user = userEvent.setup()
    await user.click(screen.getByRole('switch', { name: /强制重建索引/i }))
    await user.click(screen.getByRole('button', { name: /^启动同步$/ }))

    expect(await screen.findByText('强制重建仅允许全量全库执行，请先取消勾选目录')).toBeInTheDocument()
    expect(startCalled).toBe(false)
  })

  it('starts scoped full sync with selected roots', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')

    let inspectCalled = false
    let capturedBody: Record<string, unknown> | null = null
    mockAdminBasics({
      progress: {
        ...validProgress,
        catalogRoots: [1001, 1002, 1003],
        catalogRootNames: {
          '1001': 'A',
          '1002': 'B',
          '1003': 'C',
        },
        catalogRootProgress: {
          '1001': {
            rootFolderId: 1001,
            status: 'done',
            estimatedTotalDocs: 11,
            stats: { ...validProgress.aggregateStats, filesIndexed: 10, foldersVisited: 1 },
            updatedAt: 0,
          },
          '1002': {
            rootFolderId: 1002,
            status: 'done',
            estimatedTotalDocs: 21,
            stats: { ...validProgress.aggregateStats, filesIndexed: 20, foldersVisited: 1 },
            updatedAt: 0,
          },
          '1003': {
            rootFolderId: 1003,
            status: 'done',
            estimatedTotalDocs: 31,
            stats: { ...validProgress.aggregateStats, filesIndexed: 30, foldersVisited: 1 },
            updatedAt: 0,
          },
        },
      },
    })
    server.use(
      http.post('/npan.v1.AdminService/InspectRoots', async ({ request }) => {
        inspectCalled = true
        const body: unknown = await request.json()
        assertRecord(body)
        expect(body.folderIds).toEqual(['1001', '1002', '1003'])
        return HttpResponse.json({
          items: [
            { folderId: '1001', name: 'A', itemCount: '10', estimatedTotalDocs: '11' },
            { folderId: '1002', name: 'B', itemCount: '20', estimatedTotalDocs: '21' },
            { folderId: '1003', name: 'C', itemCount: '30', estimatedTotalDocs: '31' },
          ],
          errors: [],
        })
      }),
      http.post('/npan.v1.AdminService/StartSync', async ({ request }) => {
        const body: unknown = await request.json()
        assertRecord(body)
        capturedBody = body
        return HttpResponse.json({ message: 'Sync started' })
      }),
    )

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^启动同步$/ })).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: /刷新目录详情/i }))
    expect(inspectCalled).toBe(true)
    await user.click(screen.getByRole('button', { name: /展开/i }))
    await user.click(screen.getByRole('switch', { name: /选择根目录 1002/i }))
    await user.click(screen.getByRole('button', { name: /^启动同步$/ }))

    await waitFor(() => {
      expect(capturedBody).not.toBeNull()
    })

    const payload = getRecord(capturedBody)
    expect(payload.rootFolderIds).toEqual(['1001', '1003'])
    expect(payload.includeDepartments).toBe(false)
    expect(payload.preserveRootCatalog).toBe(true)
    expect(payload.mode).toBe('SYNC_MODE_FULL')
    expect(payload.resumeProgress).toBeUndefined()
  })

  it('keeps inspected catalog details after later progress polling overwrites base progress', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    vi.useFakeTimers({ shouldAdvanceTime: true })

    let progressRequestCount = 0
    server.use(
      http.post('/npan.v1.AdminService/GetSyncProgress', () => {
        progressRequestCount += 1
        return HttpResponse.json(toConnectProgressResponse({
          ...validProgress,
          roots: [1001, 1002],
          catalogRoots: [1001, 1002],
          rootNames: { '1001': 'A', '1002': 'B' },
          catalogRootNames: { '1001': 'A', '1002': 'B' },
          rootProgress: {},
          catalogRootProgress: {
            '1001': {
              rootFolderId: 1001,
              status: 'done',
              estimatedTotalDocs: null,
              stats: validProgress.aggregateStats,
              updatedAt: 0,
            },
            '1002': {
              rootFolderId: 1002,
              status: 'done',
              estimatedTotalDocs: null,
              stats: validProgress.aggregateStats,
              updatedAt: 0,
            },
          },
        }))
      }),
      http.post('/npan.v1.AdminService/GetIndexStats', () => {
        return HttpResponse.json({ documentCount: '10' })
      }),
      http.post('/npan.v1.AdminService/WatchSyncProgress', () => {
        return HttpResponse.json({ code: 'unimplemented' }, { status: 501 })
      }),
      http.post('/npan.v1.AdminService/InspectRoots', async ({ request }) => {
        const body: unknown = await request.json()
        assertRecord(body)
        expect(body.folderIds).toEqual(['1001', '1002'])
        return HttpResponse.json({
          items: [
            { folderId: '1001', name: 'A', itemCount: '10', estimatedTotalDocs: '11' },
            { folderId: '1002', name: 'B', itemCount: '20', estimatedTotalDocs: '21' },
          ],
          errors: [],
        })
      }),
    )

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await waitFor(() => {
      expect(progressRequestCount).toBeGreaterThan(0)
      expect(screen.getByRole('button', { name: /刷新目录详情/i })).toBeInTheDocument()
    })

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTimeAsync })
    await user.click(screen.getByRole('button', { name: /刷新目录详情/i }))
    await user.click(screen.getByRole('button', { name: /展开/i }))

    expect(await screen.findByText('估计 11')).toBeInTheDocument()
    expect(screen.getByText('估计 21')).toBeInTheDocument()

    await advanceTimers(2100)

    await waitFor(() => {
      expect(progressRequestCount).toBeGreaterThan(1)
    })
    expect(screen.getByText('估计 11')).toBeInTheDocument()
    expect(screen.getByText('估计 21')).toBeInTheDocument()
  })

  it('keeps inspected catalog details when later progress requests return not found', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    vi.useFakeTimers({ shouldAdvanceTime: true })

    let progressRequestCount = 0
    server.use(
      http.post('/npan.v1.AdminService/GetSyncProgress', () => {
        progressRequestCount += 1
        if (progressRequestCount === 1) {
          return HttpResponse.json(toConnectProgressResponse({
            ...validProgress,
            roots: [1001],
            catalogRoots: [1001],
            rootNames: { '1001': 'A' },
            catalogRootNames: { '1001': 'A' },
            rootProgress: {},
            catalogRootProgress: {
              '1001': {
                rootFolderId: 1001,
                status: 'done',
                estimatedTotalDocs: null,
                stats: validProgress.aggregateStats,
                updatedAt: 0,
              },
            },
          }))
        }
        return HttpResponse.json({ code: 'not_found', message: '未找到同步进度' }, { status: 404 })
      }),
      http.post('/npan.v1.AdminService/GetIndexStats', () => {
        return HttpResponse.json({ documentCount: '10' })
      }),
      http.post('/npan.v1.AdminService/WatchSyncProgress', () => {
        return HttpResponse.json({ code: 'unimplemented' }, { status: 501 })
      }),
      http.post('/npan.v1.AdminService/InspectRoots', () => {
        return HttpResponse.json({
          items: [
            { folderId: '1001', name: 'A', itemCount: '10', estimatedTotalDocs: '11' },
          ],
          errors: [],
        })
      }),
    )

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await waitFor(() => {
      expect(progressRequestCount).toBe(1)
      expect(screen.getByRole('button', { name: /刷新目录详情/i })).toBeInTheDocument()
    })

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTimeAsync })
    await user.click(screen.getByRole('button', { name: /刷新目录详情/i }))
    await user.click(screen.getByRole('button', { name: /展开/i }))

    expect(await screen.findByText('估计 11')).toBeInTheDocument()

    await advanceTimers(2100)

    await waitFor(() => {
      expect(progressRequestCount).toBeGreaterThan(1)
    })
    expect(screen.queryByText('暂无同步记录')).not.toBeInTheDocument()
    expect(screen.getByText('估计 11')).toBeInTheDocument()
  })

  it('full sync with force rebuild explicitly disables resume', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')

    let capturedBody: Record<string, unknown> | null = null
    mockAdminBasics({ progress: validProgress })
    server.use(
      http.post('/npan.v1.AdminService/StartSync', async ({ request }) => {
        const body: unknown = await request.json()
        assertRecord(body)
        capturedBody = body
        return HttpResponse.json({ message: 'Sync started' })
      }),
    )

    const isolatedWrapper = createTestProvider()
    render(<AdminSyncPage />, { wrapper: isolatedWrapper })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^启动同步$/ })).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.click(screen.getByRole('switch', { name: /强制重建索引/i }))
    await user.click(screen.getByRole('button', { name: /^启动同步$/ }))
    await user.click(screen.getByRole('button', { name: '确认重建' }))

    await waitFor(() => {
      expect(capturedBody).not.toBeNull()
    })

    const payload = getRecord(capturedBody)
    expect(payload.mode).toBe('SYNC_MODE_FULL')
    expect(payload.forceRebuild).toBe(true)
    expect(payload.resumeProgress).toBe(false)
  })
})
