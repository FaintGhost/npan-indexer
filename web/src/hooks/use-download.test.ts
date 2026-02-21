import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { useDownload } from './use-download'

describe('useDownload', () => {
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
      http.get('/api/v1/app/download-url', ({ request }) => {
        const url = new URL(request.url)
        expect(url.searchParams.get('file_id')).toBe('42')
        return HttpResponse.json({
          file_id: 42,
          download_url: 'https://cdn.example.com/file.pdf',
        })
      }),
    )

    const { result } = renderHook(() => useDownload())

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
      http.get('/api/v1/app/download-url', async () => {
        await new Promise((r) => setTimeout(r, 1000))
        return HttpResponse.json({
          file_id: 42,
          download_url: 'https://cdn.example.com/file.pdf',
        })
      }),
    )

    const { result } = renderHook(() => useDownload())

    act(() => {
      result.current.download(42)
    })

    expect(result.current.getStatus(42)).toBe('loading')
  })

  it('reverts to idle after success timeout', async () => {
    server.use(
      http.get('/api/v1/app/download-url', () => {
        return HttpResponse.json({
          file_id: 42,
          download_url: 'https://cdn.example.com/file.pdf',
        })
      }),
    )

    const { result } = renderHook(() => useDownload())

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
      http.get('/api/v1/app/download-url', () => {
        requestCount++
        return HttpResponse.json({
          file_id: 42,
          download_url: 'https://cdn.example.com/file.pdf',
        })
      }),
    )

    const { result } = renderHook(() => useDownload())

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
      http.get('/api/v1/app/download-url', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed' },
          { status: 502 },
        )
      }),
    )

    const { result } = renderHook(() => useDownload())

    await act(async () => {
      result.current.download(42)
    })

    await waitFor(() => {
      expect(result.current.getStatus(42)).toBe('error')
    })
  })

  it('returns idle for unknown file ids', () => {
    const { result } = renderHook(() => useDownload())
    expect(result.current.getStatus(999)).toBe('idle')
  })
})
