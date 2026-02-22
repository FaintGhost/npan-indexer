import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { useSyncProgress } from './use-sync-progress'

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

  describe('startSync 后状态自动刷新', () => {
    const doneProgress = {
      ...validProgress,
      status: 'done' as const,
      startedAt: 1000,
      updatedAt: 2000,
      roots: [100],
      completedRoots: [100],
      aggregateStats: {
        ...validProgress.aggregateStats,
        filesIndexed: 42,
      },
    }

    it('startSync 后 progress 立即变为 running（乐观更新）', async () => {
      let getCalls = 0
      server.use(
        http.get('/api/v1/admin/sync', () => {
          getCalls++
          return HttpResponse.json(doneProgress)
        }),
        http.post('/api/v1/admin/sync', () => {
          return HttpResponse.json({ message: 'ok' }, { status: 202 })
        }),
      )

      const { result } = renderHook(() => useSyncProgress(headers))

      // Wait for initial fetch
      await waitFor(() => {
        expect(result.current.progress).not.toBeNull()
      })

      // startSync should set progress.status to "running" optimistically,
      // even though GET returns "done" (old data)
      await act(async () => {
        await result.current.startSync([100], 'auto')
      })

      expect(result.current.progress?.status).toBe('running')
    })

    it('startSync 后轮询不因旧数据停止（宽限期内）', async () => {
      let getCalls = 0
      server.use(
        http.get('/api/v1/admin/sync', () => {
          getCalls++
          return HttpResponse.json(doneProgress)
        }),
        http.post('/api/v1/admin/sync', () => {
          return HttpResponse.json({ message: 'ok' }, { status: 202 })
        }),
      )

      const { result } = renderHook(() => useSyncProgress(headers))

      // Wait for initial fetch
      await waitFor(() => {
        expect(result.current.progress).not.toBeNull()
      })

      const callsBeforeSync = getCalls

      await act(async () => {
        await result.current.startSync([100], 'auto')
      })

      // Record calls right after startSync (includes the fetchProgress in startSync)
      const callsAfterSync = getCalls

      // Advance past one poll interval — polling should still be active
      // even though GET keeps returning "done" (grace period)
      await act(async () => {
        vi.advanceTimersByTime(2000)
      })

      const callsAfterFirstPoll = getCalls
      expect(callsAfterFirstPoll).toBeGreaterThan(callsAfterSync)

      // Advance another poll interval — still within grace period
      await act(async () => {
        vi.advanceTimersByTime(2000)
      })

      const callsAfterSecondPoll = getCalls
      expect(callsAfterSecondPoll).toBeGreaterThan(callsAfterFirstPoll)
    })

    it('宽限期结束后轮询正常停止', async () => {
      let getCalls = 0
      server.use(
        http.get('/api/v1/admin/sync', () => {
          getCalls++
          return HttpResponse.json(doneProgress)
        }),
        http.post('/api/v1/admin/sync', () => {
          return HttpResponse.json({ message: 'ok' }, { status: 202 })
        }),
      )

      const { result } = renderHook(() => useSyncProgress(headers))

      // Wait for initial fetch
      await waitFor(() => {
        expect(result.current.progress).not.toBeNull()
      })

      await act(async () => {
        await result.current.startSync([100], 'auto')
      })

      // Advance past the grace period (5 polls × 2000ms = 10000ms + extra)
      await act(async () => {
        vi.advanceTimersByTime(12000)
      })

      // Record the call count after grace period expires
      const callsAfterGrace = getCalls

      // Advance another poll interval — polling should have stopped
      await act(async () => {
        vi.advanceTimersByTime(4000)
      })

      const callsAfterExtra = getCalls
      expect(callsAfterExtra).toBe(callsAfterGrace)
    })
  })

})
