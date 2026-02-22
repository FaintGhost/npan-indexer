import { describe, it, expect, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { useAdminAuth } from './use-admin-auth'

const STORAGE_KEY = 'npan_admin_api_key'

describe('useAdminAuth', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('needsAuth is true when no stored key', () => {
    const { result } = renderHook(() => useAdminAuth())
    expect(result.current.needsAuth).toBe(true)
  })

  it('needsAuth is false when key is stored', () => {
    localStorage.setItem(STORAGE_KEY, 'test-key')
    const { result } = renderHook(() => useAdminAuth())
    expect(result.current.needsAuth).toBe(false)
  })

  it('validates and stores valid key', async () => {
    server.use(
      http.get('/api/v1/admin/sync', ({ request }) => {
        const key = request.headers.get('X-API-Key')
        if (key === 'valid-key') {
          return HttpResponse.json({
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
          })
        }
        return HttpResponse.json(
          { code: 'UNAUTHORIZED', message: 'Invalid API key' },
          { status: 401 },
        )
      }),
    )

    const { result } = renderHook(() => useAdminAuth())

    await act(async () => {
      const ok = await result.current.validate('valid-key')
      expect(ok).toBe(true)
    })

    expect(result.current.needsAuth).toBe(false)
    expect(localStorage.getItem(STORAGE_KEY)).toBe('valid-key')
  })

  it('rejects invalid key', async () => {
    server.use(
      http.get('/api/v1/admin/sync', () => {
        return HttpResponse.json(
          { code: 'UNAUTHORIZED', message: 'Invalid' },
          { status: 401 },
        )
      }),
    )

    const { result } = renderHook(() => useAdminAuth())

    await act(async () => {
      const ok = await result.current.validate('bad-key')
      expect(ok).toBe(false)
    })

    expect(result.current.needsAuth).toBe(true)
    expect(result.current.error).toBeTruthy()
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull()
  })

  it('rejects empty key without request', async () => {
    const { result } = renderHook(() => useAdminAuth())

    await act(async () => {
      const ok = await result.current.validate('')
      expect(ok).toBe(false)
    })

    expect(result.current.error).toBeTruthy()
  })

  it('on401 clears storage and sets needsAuth', () => {
    localStorage.setItem(STORAGE_KEY, 'old-key')
    const { result } = renderHook(() => useAdminAuth())
    expect(result.current.needsAuth).toBe(false)

    act(() => {
      result.current.on401()
    })

    expect(result.current.needsAuth).toBe(true)
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull()
  })

  it('getHeaders returns X-API-Key header', () => {
    localStorage.setItem(STORAGE_KEY, 'my-key')
    const { result } = renderHook(() => useAdminAuth())
    expect(result.current.getHeaders()).toEqual({ 'X-API-Key': 'my-key' })
  })
})
