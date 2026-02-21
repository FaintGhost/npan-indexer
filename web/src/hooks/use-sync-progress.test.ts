import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { useSyncProgress } from './use-sync-progress'

const validProgress = {
  status: 'idle',
  startedAt: 0,
  updatedAt: 0,
  roots: [],
  completedRoots: [],
  aggregateStats: {
    foldersVisited: 0,
    filesIndexed: 0,
    pagesFetched: 0,
    failedRequests: 0,
    startedAt: 0,
    endedAt: 0,
  },
  rootProgress: {},
}

describe('useSyncProgress', () => {
  const headers = { 'X-API-Key': 'test-key' }

  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('fetches initial progress', async () => {
    server.use(
      http.get('/api/v1/admin/sync', () => {
        return HttpResponse.json({
          ...validProgress,
          status: 'done',
          roots: [100, 200],
          completedRoots: [100, 200],
          aggregateStats: { ...validProgress.aggregateStats, filesIndexed: 500 },
        })
      }),
    )

    const { result } = renderHook(() => useSyncProgress(headers))

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull()
      expect(result.current.progress?.status).toBe('done')
      expect(result.current.progress?.aggregateStats.filesIndexed).toBe(500)
    })
  })

  it('starts sync', async () => {
    let postCalled = false
    server.use(
      http.get('/api/v1/admin/sync', () => {
        return HttpResponse.json(validProgress)
      }),
      http.post('/api/v1/admin/sync', () => {
        postCalled = true
        return HttpResponse.json({ message: 'Sync started' })
      }),
    )

    const { result } = renderHook(() => useSyncProgress(headers))

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull()
    })

    await act(async () => {
      await result.current.startSync([100, 200])
    })

    expect(postCalled).toBe(true)
  })

  it('cancels sync', async () => {
    let cancelCalled = false
    server.use(
      http.get('/api/v1/admin/sync', () => {
        return HttpResponse.json({ ...validProgress, status: 'running' })
      }),
      http.delete('/api/v1/admin/sync', () => {
        cancelCalled = true
        return HttpResponse.json({ message: 'Cancelled' })
      }),
    )

    const { result } = renderHook(() => useSyncProgress(headers))

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull()
    })

    await act(async () => {
      await result.current.cancelSync()
    })

    expect(cancelCalled).toBe(true)
  })

  it('sets error on failed request', async () => {
    server.use(
      http.get('/api/v1/admin/sync', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Server error' },
          { status: 500 },
        )
      }),
    )

    const { result } = renderHook(() => useSyncProgress(headers))

    await waitFor(() => {
      expect(result.current.error).toBeTruthy()
    })
  })
})
