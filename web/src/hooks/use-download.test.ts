import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { createTestProvider } from '../tests/test-providers'
import { useDownload } from './use-download'

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

describe('useDownload', () => {
  const wrapper = createTestProvider()
  const mockOpen = vi.fn()

  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    vi.stubGlobal('open', mockOpen)
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('fetches download URL and opens in new tab', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppDownloadURL', async ({ request }) => {
        const body: unknown = await request.json()
        expect(isRecord(body)).toBe(true)
        if (!isRecord(body)) {
          throw new Error('expected download request body to be an object')
        }
        expect(body).toEqual({ fileId: '42' })
        return HttpResponse.json({
          result: {
            fileId: '42',
            downloadUrl: 'https://cdn.example.com/file.pdf',
          },
        })
      }),
    )

    const { result } = renderHook(() => useDownload(), { wrapper })

    await act(async () => {
      result.current.download(42)
    })

    await waitFor(() => {
      expect(result.current.getStatus(42)).toBe('success')
    })

    expect(mockOpen).toHaveBeenCalledWith(
      'https://cdn.example.com/file.pdf',
      '_blank',
      'noopener,noreferrer',
    )
  })

  it('shows loading state during request', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppDownloadURL', async () => {
        await new Promise((r) => setTimeout(r, 1000))
        return HttpResponse.json({
          result: {
            fileId: '42',
            downloadUrl: 'https://cdn.example.com/file.pdf',
          },
        })
      }),
    )

    const { result } = renderHook(() => useDownload(), { wrapper })

    act(() => {
      result.current.download(42)
    })

    expect(result.current.getStatus(42)).toBe('loading')
  })

  it('reverts to idle after success timeout', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppDownloadURL', () => {
        return HttpResponse.json({
          result: {
            fileId: '42',
            downloadUrl: 'https://cdn.example.com/file.pdf',
          },
        })
      }),
    )

    const { result } = renderHook(() => useDownload(), { wrapper })

    await act(async () => {
      result.current.download(42)
    })

    await waitFor(() => {
      expect(result.current.getStatus(42)).toBe('success')
    })

    act(() => {
      vi.advanceTimersByTime(2000)
    })

    expect(result.current.getStatus(42)).toBe('idle')
  })

  it('uses cached URL on second download', async () => {
    let requestCount = 0
    server.use(
      http.post('/npan.v1.AppService/AppDownloadURL', () => {
        requestCount++
        return HttpResponse.json({
          result: {
            fileId: '42',
            downloadUrl: 'https://cdn.example.com/file.pdf',
          },
        })
      }),
    )

    const { result } = renderHook(() => useDownload(), { wrapper })

    await act(async () => {
      result.current.download(42)
    })

    await waitFor(() => {
      expect(result.current.getStatus(42)).toBe('success')
    })

    // Reset status
    act(() => {
      vi.advanceTimersByTime(2000)
    })

    // Second download should use cache
    await act(async () => {
      result.current.download(42)
    })

    expect(requestCount).toBe(1) // Only 1 API call
    expect(mockOpen).toHaveBeenCalledTimes(2) // But opened twice
  })

  it('shows error state on API failure', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppDownloadURL', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed' },
          { status: 502 },
        )
      }),
    )

    const { result } = renderHook(() => useDownload(), { wrapper })

    await act(async () => {
      result.current.download(42)
    })

    await waitFor(() => {
      expect(result.current.getStatus(42)).toBe('error')
    })
  })

  it('returns idle for unknown file ids', () => {
    const { result } = renderHook(() => useDownload(), { wrapper })
    expect(result.current.getStatus(999)).toBe('idle')
  })
})
