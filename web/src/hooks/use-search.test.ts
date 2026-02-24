import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { createTestProvider } from '../tests/test-providers'
import { useSearch } from './use-search'

type SearchItem = {
  doc_id: string
  source_id: number
  type: 'file' | 'folder'
  name: string
  path_text: string
  parent_id: number
  modified_at: number
  created_at: number
  size: number
  sha1: string
  in_trash: boolean
  is_deleted: boolean
  highlighted_name?: string
}

function assertRecord(value: unknown): asserts value is Record<string, unknown> {
  if (typeof value !== 'object' || value === null) {
    throw new Error('expected payload to be an object')
  }
}

function mockSearchResponse(
  items: Array<{ doc_id: string; source_id: number; name: string }>,
): SearchItem[] {
  return items.map((item) => ({
    doc_id: item.doc_id,
    source_id: item.source_id,
    type: 'file',
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

function toConnectSearchResponse(items: SearchItem[], total: number) {
  return {
    result: {
      items: items.map((item) => ({
        docId: item.doc_id,
        sourceId: String(item.source_id),
        type: item.type === 'folder' ? 'ITEM_TYPE_FOLDER' : 'ITEM_TYPE_FILE',
        name: item.name,
        pathText: item.path_text,
        parentId: String(item.parent_id),
        modifiedAt: String(item.modified_at),
        createdAt: String(item.created_at),
        size: String(item.size),
        sha1: item.sha1,
        inTrash: item.in_trash,
        isDeleted: item.is_deleted,
        highlightedName: item.highlighted_name,
      })),
      total: String(total),
    },
  }
}

describe('useSearch', () => {
  const wrapper = createTestProvider()

  function renderSearchHook() {
    return renderHook(() => useSearch(), { wrapper })
  }

  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('searches after debounce delay', async () => {
    const items = mockSearchResponse([{ doc_id: 'f1', source_id: 1, name: 'MX40.pdf' }])

    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(toConnectSearchResponse(items, 1))
      }),
    )

    const { result } = renderSearchHook()

    act(() => {
      result.current.setQuery('MX40')
    })

    expect(result.current.items).toHaveLength(0)

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
      http.post('/npan.v1.AppService/AppSearch', async ({ request }) => {
        requestCount++
        const body: unknown = await request.json()
        assertRecord(body)
        const q = String(body.query ?? '')
        return HttpResponse.json(
          toConnectSearchResponse(
            mockSearchResponse([{ doc_id: 'f1', source_id: 1, name: `${q}.pdf` }]),
            1,
          ),
        )
      }),
    )

    const { result } = renderSearchHook()

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

    expect(requestCount).toBe(1)
  })

  it('immediately searches on searchImmediate', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(
          toConnectSearchResponse(
            mockSearchResponse([{ doc_id: 'f1', source_id: 1, name: '固件.bin' }]),
            1,
          ),
        )
      }),
    )

    const { result } = renderSearchHook()

    await act(async () => {
      await result.current.searchImmediate('固件')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1)
    })
  })

  it('loads more pages', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppSearch', async ({ request }) => {
        const body: unknown = await request.json()
        assertRecord(body)
        const p = Number(body.page ?? '1')
        const items = mockSearchResponse(
          Array.from({ length: 3 }, (_, i) => ({
            doc_id: `f${(p - 1) * 3 + i + 1}`,
            source_id: (p - 1) * 3 + i + 1,
            name: `file${(p - 1) * 3 + i + 1}.pdf`,
          })),
        )
        return HttpResponse.json(toConnectSearchResponse(items, 10))
      }),
    )

    const { result } = renderSearchHook()

    await act(async () => {
      await result.current.searchImmediate('test')
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
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(
          toConnectSearchResponse(
            mockSearchResponse([{ doc_id: 'f1', source_id: 1, name: 'only.pdf' }]),
            1,
          ),
        )
      }),
    )

    const { result } = renderSearchHook()

    await act(async () => {
      await result.current.searchImmediate('only')
    })

    await waitFor(() => {
      expect(result.current.hasMore).toBe(false)
    })
  })

  it('deduplicates items by source_id', async () => {
    let callCount = 0
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        callCount++
        if (callCount === 1) {
          return HttpResponse.json(
            toConnectSearchResponse(
              mockSearchResponse([
                { doc_id: 'f1', source_id: 1, name: 'a.pdf' },
                { doc_id: 'f2', source_id: 2, name: 'b.pdf' },
                { doc_id: 'f3', source_id: 3, name: 'c.pdf' },
              ]),
              6,
            ),
          )
        }
        return HttpResponse.json(
          toConnectSearchResponse(
            mockSearchResponse([
              { doc_id: 'f3', source_id: 3, name: 'c.pdf' },
              { doc_id: 'f4', source_id: 4, name: 'd.pdf' },
              { doc_id: 'f5', source_id: 5, name: 'e.pdf' },
            ]),
            6,
          ),
        )
      }),
    )

    const { result } = renderSearchHook()

    await act(async () => {
      await result.current.searchImmediate('test')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(3)
    })

    await act(async () => {
      result.current.loadMore()
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(5)
    })
  })

  it('resets state', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(
          toConnectSearchResponse(
            mockSearchResponse([{ doc_id: 'f1', source_id: 1, name: 'test.pdf' }]),
            1,
          ),
        )
      }),
    )

    const { result } = renderSearchHook()

    await act(async () => {
      await result.current.searchImmediate('test')
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
      http.post('/npan.v1.AppService/AppSearch', async () => {
        await new Promise((r) => setTimeout(r, 100))
        return HttpResponse.json(toConnectSearchResponse([], 0))
      }),
    )

    const { result } = renderSearchHook()

    act(() => {
      void result.current.searchImmediate('test')
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
      http.post('/npan.v1.AppService/AppSearch', () => {
        callCount++
        return HttpResponse.json(
          toConnectSearchResponse(
            mockSearchResponse([
              {
                doc_id: `f${callCount}`,
                source_id: callCount,
                name: `file${callCount}.pdf`,
              },
            ]),
            1,
          ),
        )
      }),
    )

    const { result } = renderSearchHook()

    await act(async () => {
      await result.current.searchImmediate('first')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1)
      expect(result.current.items[0]?.name).toBe('file1.pdf')
    })

    await act(async () => {
      await result.current.searchImmediate('second')
    })

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1)
      expect(result.current.items[0]?.name).toBe('file2.pdf')
    })
  })
})
