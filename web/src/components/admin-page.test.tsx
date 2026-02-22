import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { AdminSyncPage } from './admin-sync-page'

const STORAGE_KEY = 'npan_admin_api_key'

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
      expect(screen.getByText(/启动同步/)).toBeInTheDocument()
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
      expect(screen.getByText(/启动同步/)).toBeInTheDocument()
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
})
