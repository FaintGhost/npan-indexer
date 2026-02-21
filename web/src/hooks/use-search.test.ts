import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { useSearch } from './use-search'

function mockSearchResponse(items: Array<{ doc_id: string; source_id: number; name: string }>, _total: number) {
  return items.map((item) => ({
    doc_id: item.doc_id,
    source_id: item.source_id,
    type: 'file' as const,
    name: item.name,
    path_text: `/${item.name}`,
    parent_id: 0,
    modified_at: 1700000000,
    created_at: 1700000000,
    size: 1024,
    sha1: 'abc',
    in_trash: false,
    is_deleted: false,
    highlighted_name: '',
  }))
}

describe('useSearch', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('searches after debounce delay', async () => {
    const items = mockSearchResponse([
      { doc_id: 'f1', source_id: 1, name: 'MX40.pdf' },
    ], 1)

    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json({ items, total: 1 })
      }),
    )

    const { result } = renderHook(() => useSearch())

    act(() => {
      result.current.setQuery('MX40')
    })

    // Before debounce fires
    expect(result.current.items).toHaveLength(0)

    // Advance past debounce (280ms)
    await act(async () => {
      vi.advanceTimersByTime(300)
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1)
      expect(result.current.total).toBe(1)
    })
  })

  it('coalesces rapid inputs', async () => {
    let requestCount = 0

    server.use(
      http.get('/api/v1/app/search', ({ request }) => {
        requestCount++
        const url = new URL(request.url)
        const q = url.searchParams.get('query')
        return HttpResponse.json({
          items: mockSearchResponse([
            { doc_id: 'f1', source_id: 1, name: `${q}.pdf` },
          ], 1),
          total: 1,
        })
      }),
    )

    const { result } = renderHook(() => useSearch())

    act(() => {
      result.current.setQuery('M')
    })
    act(() => {
      vi.advanceTimersByTime(50)
    })
    act(() => {
      result.current.setQuery('MX')
    })
    act(() => {
      vi.advanceTimersByTime(50)
    })
    act(() => {
      result.current.setQuery('MX40')
    })

    await act(async () => {
      vi.advanceTimersByTime(300)
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1)
    })

    // Should only have made 1 request (for "MX40")
    expect(requestCount).toBe(1)
  })

  it('immediately searches on searchImmediate', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json({
          items: mockSearchResponse([
            { doc_id: 'f1', source_id: 1, name: '固件.bin' },
          ], 1),
          total: 1,
        })
      }),
    )

    const { result } = renderHook(() => useSearch())

    await act(async () => {
      result.current.searchImmediate('固件')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1)
    })
  })

  it('loads more pages', async () => {
    let page = 0
    server.use(
      http.get('/api/v1/app/search', ({ request }) => {
        page++
        const url = new URL(request.url)
        const p = Number(url.searchParams.get('page') || 1)
        const items = mockSearchResponse(
          Array.from({ length: 3 }, (_, i) => ({
            doc_id: `f${(p - 1) * 3 + i + 1}`,
            source_id: (p - 1) * 3 + i + 1,
            name: `file${(p - 1) * 3 + i + 1}.pdf`,
          })),
          10,
        )
        return HttpResponse.json({ items, total: 10 })
      }),
    )

    const { result } = renderHook(() => useSearch())

    await act(async () => {
      result.current.searchImmediate('test')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(3)
      expect(result.current.hasMore).toBe(true)
    })

    await act(async () => {
      result.current.loadMore()
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(6)
    })
  })

  it('stops loading when no more pages', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json({
          items: mockSearchResponse([
            { doc_id: 'f1', source_id: 1, name: 'only.pdf' },
          ], 1),
          total: 1,
        })
      }),
    )

    const { result } = renderHook(() => useSearch())

    await act(async () => {
      result.current.searchImmediate('only')
    })

    await waitFor(() => {
      expect(result.current.hasMore).toBe(false)
    })
  })

  it('deduplicates items by source_id', async () => {
    let callCount = 0
    server.use(
      http.get('/api/v1/app/search', () => {
        callCount++
        if (callCount === 1) {
          return HttpResponse.json({
            items: mockSearchResponse([
              { doc_id: 'f1', source_id: 1, name: 'a.pdf' },
              { doc_id: 'f2', source_id: 2, name: 'b.pdf' },
              { doc_id: 'f3', source_id: 3, name: 'c.pdf' },
            ], 6),
            total: 6,
          })
        }
        return HttpResponse.json({
          items: mockSearchResponse([
            { doc_id: 'f3', source_id: 3, name: 'c.pdf' },
            { doc_id: 'f4', source_id: 4, name: 'd.pdf' },
            { doc_id: 'f5', source_id: 5, name: 'e.pdf' },
          ], 6),
          total: 6,
        })
      }),
    )

    const { result } = renderHook(() => useSearch())

    await act(async () => {
      result.current.searchImmediate('test')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(3)
    })

    await act(async () => {
      result.current.loadMore()
    })

    await waitFor(() => {
      // 5 unique items (source_id 3 appears in both pages)
      expect(result.current.items).toHaveLength(5)
    })
  })

  it('resets state', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json({
          items: mockSearchResponse([
            { doc_id: 'f1', source_id: 1, name: 'test.pdf' },
          ], 1),
          total: 1,
        })
      }),
    )

    const { result } = renderHook(() => useSearch())

    await act(async () => {
      result.current.searchImmediate('test')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1)
    })

    act(() => {
      result.current.reset()
    })

    expect(result.current.items).toHaveLength(0)
    expect(result.current.total).toBe(0)
    expect(result.current.query).toBe('')
  })

  it('sets loading state during search', async () => {
    server.use(
      http.get('/api/v1/app/search', async () => {
        await new Promise((r) => setTimeout(r, 100))
        return HttpResponse.json({ items: [], total: 0 })
      }),
    )

    const { result } = renderHook(() => useSearch())

    act(() => {
      result.current.searchImmediate('test')
    })

    expect(result.current.loading).toBe(true)

    await act(async () => {
      vi.advanceTimersByTime(200)
    })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
  })

  it('clears items on new query', async () => {
    let callCount = 0
    server.use(
      http.get('/api/v1/app/search', () => {
        callCount++
        return HttpResponse.json({
          items: mockSearchResponse([
            { doc_id: `f${callCount}`, source_id: callCount, name: `file${callCount}.pdf` },
          ], 1),
          total: 1,
        })
      }),
    )

    const { result } = renderHook(() => useSearch())

    await act(async () => {
      result.current.searchImmediate('first')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1)
    })

    await act(async () => {
      result.current.searchImmediate('second')
    })

    await waitFor(() => {
      // New query should have replaced old items
      expect(result.current.items).toHaveLength(1)
    })
  })
})
