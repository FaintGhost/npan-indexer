import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
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

const validProgress = {
  status: 'idle',
  startedAt: 0,
  updatedAt: 0,
  meiliHost: 'http://localhost:7700',
  meiliIndex: 'documents',
  checkpointTemplate: '',
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

describe('AdminSyncPage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('shows API key dialog when no stored key', () => {
    render(<AdminSyncPage />)
    expect(screen.getByPlaceholderText(/API Key/i)).toBeInTheDocument()
  })

  it('shows admin panel when key is stored', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    server.use(
      http.get('/api/v1/admin/sync', () => {
        return HttpResponse.json(validProgress)
      }),
    )

    render(<AdminSyncPage />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^启动同步$/ })).toBeInTheDocument()
    })
  })

  it('closes dialog after valid key input', async () => {
    server.use(
      http.get('/api/v1/admin/sync', ({ request }) => {
        const key = request.headers.get('X-API-Key')
        if (key === 'valid-key') {
          return HttpResponse.json(validProgress)
        }
        return HttpResponse.json(
          { code: 'UNAUTHORIZED', message: 'Invalid' },
          { status: 401 },
        )
      }),
    )

    render(<AdminSyncPage />)

    const user = userEvent.setup()
    await user.type(screen.getByPlaceholderText(/API Key/i), 'valid-key')
    await user.click(screen.getByRole('button', { name: /确认/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^启动同步$/ })).toBeInTheDocument()
    })
  })

  it('shows progress when sync is running', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')
    server.use(
      http.get('/api/v1/admin/sync', () => {
        return HttpResponse.json({
          ...validProgress,
          status: 'running',
          roots: [100, 200],
          completedRoots: [100],
          aggregateStats: {
            ...validProgress.aggregateStats,
            filesIndexed: 300,
          },
        })
      }),
    )

    render(<AdminSyncPage />)

    await waitFor(() => {
      expect(screen.getByText('运行中')).toBeInTheDocument()
      expect(screen.getByText(/300/)).toBeInTheDocument()
    })
  })

  it('inspects folders first, then starts scoped full sync from selected roots', async () => {
    localStorage.setItem(STORAGE_KEY, 'valid-key')

    let inspectCalled = false
    let capturedBody: Record<string, unknown> | null = null
    server.use(
      http.get('/api/v1/admin/sync', () => {
        return HttpResponse.json({
          ...validProgress,
          catalogRoots: [1001, 1002, 1003],
          catalogRootNames: {
            1001: 'A',
            1002: 'B',
            1003: 'C',
          },
          catalogRootProgress: {
            '1001': {
              rootFolderId: 1001,
              status: 'done',
              estimatedTotalDocs: 11,
              stats: {
                foldersVisited: 1,
                filesIndexed: 10,
                filesDiscovered: 10,
                skippedFiles: 0,
                pagesFetched: 1,
                failedRequests: 0,
                startedAt: 0,
                endedAt: 0,
              },
              updatedAt: 0,
            },
            '1002': {
              rootFolderId: 1002,
              status: 'done',
              estimatedTotalDocs: 21,
              stats: {
                foldersVisited: 1,
                filesIndexed: 20,
                filesDiscovered: 20,
                skippedFiles: 0,
                pagesFetched: 1,
                failedRequests: 0,
                startedAt: 0,
                endedAt: 0,
              },
              updatedAt: 0,
            },
            '1003': {
              rootFolderId: 1003,
              status: 'done',
              estimatedTotalDocs: 31,
              stats: {
                foldersVisited: 1,
                filesIndexed: 30,
                filesDiscovered: 30,
                skippedFiles: 0,
                pagesFetched: 1,
                failedRequests: 0,
                startedAt: 0,
                endedAt: 0,
              },
              updatedAt: 0,
            },
          },
        })
      }),
      http.post('/api/v1/admin/roots/inspect', async ({ request }) => {
        inspectCalled = true
        const body: unknown = await request.json()
        assertRecord(body)
        expect(body.folder_ids).toEqual([1001, 1002, 1003])
        return HttpResponse.json({
          items: [
            { folder_id: 1001, name: 'A', item_count: 10, estimated_total_docs: 11 },
            { folder_id: 1002, name: 'B', item_count: 20, estimated_total_docs: 21 },
            { folder_id: 1003, name: 'C', item_count: 30, estimated_total_docs: 31 },
          ],
          errors: [],
        })
      }),
      http.post('/api/v1/admin/sync', async ({ request }) => {
        const body: unknown = await request.json()
        assertRecord(body)
        capturedBody = body
        return HttpResponse.json({ message: 'Sync started' }, { status: 202 })
      }),
    )

    render(<AdminSyncPage />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^启动同步$/ })).toBeInTheDocument()
    })

    const user = userEvent.setup()
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: /刷新目录详情/i }))
    expect(inspectCalled).toBe(true)

    await user.click(screen.getByRole('button', { name: /^全量$/ }))
    await user.click(screen.getByRole('button', { name: /展开/i }))
    await user.click(screen.getByRole('switch', { name: /选择根目录 1002/i }))

    await user.click(screen.getByRole('button', { name: /按勾选目录启动全量/i }))

    await waitFor(() => {
      expect(capturedBody).not.toBeNull()
    })

    const payload = getRecord(capturedBody)
    expect(payload.root_folder_ids).toEqual([1001, 1003])
    expect(payload.include_departments).toBe(false)
    expect(payload.preserve_root_catalog).toBe(true)
  })
})
