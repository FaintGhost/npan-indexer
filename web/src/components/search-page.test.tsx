import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { createTestProvider } from '../tests/test-providers'
import { SearchPage } from '../routes/index.lazy'
import '../app.css'

const { createPublicSearchClientSpy } = vi.hoisted(() => ({
  createPublicSearchClientSpy: vi.fn(),
}))

vi.mock('@/lib/meili-search-client', () => ({
  createPublicSearchClient: createPublicSearchClientSpy,
}))

// Helper to create search response
const GET_SEARCH_CONFIG_PATH = '/npan.v1.AppService/GetSearchConfig'

function makeSearchResponse(count: number, total: number) {
  return {
    items: Array.from({ length: count }, (_, i) => ({
      doc_id: `f${i + 1}`,
      source_id: i + 1,
      type: 'file',
      name: `file${i + 1}.pdf`,
      path_text: `/file${i + 1}.pdf`,
      parent_id: 0,
      modified_at: 1700000000,
      created_at: 1700000000,
      size: 1024 * (i + 1),
      sha1: `hash${i}`,
      in_trash: false,
      is_deleted: false,
      highlighted_name: '',
    })),
    total,
  }
}

function makeSearchConfigResponse(overrides?: {
  host?: string
  indexName?: string
  searchApiKey?: string
  instantsearchEnabled?: boolean
}) {
  return {
    host: 'https://search.example.com',
    indexName: 'npan-public',
    searchApiKey: 'public-search-key',
    instantsearchEnabled: true,
    ...overrides,
  }
}

function toConnectSearchResponse(items: ReturnType<typeof makeSearchResponse>['items'], total: number) {
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

function makeInstantSearchClient(
  hits: Array<Record<string, unknown>>,
  total: number,
  options?: {
    onSearch?: (requests: Array<{ params?: Record<string, unknown> }>) => void
    resolveResult?: (
      requests: Array<{ params?: Record<string, unknown> }>,
    ) => { hits: Array<Record<string, unknown>>; total: number }
  },
) {
  return {
    searchClient: {
      search: async (requests: Array<{ params?: Record<string, unknown> }>) => {
        const clonedRequests = requests.map((request) => ({
          ...request,
          params: request.params ? { ...request.params } : undefined,
        }))

        options?.onSearch?.(clonedRequests)

        const resolved = options?.resolveResult?.(clonedRequests) ?? { hits, total }
        const requestQuery = typeof requests[0]?.params?.query === 'string'
          ? requests[0].params.query
          : ''

        return {
          results: [
            {
              hits: resolved.hits,
              nbHits: resolved.total,
              page: 0,
              nbPages: resolved.total > resolved.hits.length ? 2 : 1,
              hitsPerPage: resolved.hits.length || 20,
              processingTimeMS: 1,
              query: requestQuery,
              params: '',
              exhaustiveNbHits: true,
            },
          ],
        }
      },
    },
    setMeiliSearchParams: () => {},
    meiliSearchInstance: {},
  }
}

function setTestURL(path: string) {
  window.history.pushState({}, '', path)
}

function makeItem(sourceID: number, name: string) {
  return {
    doc_id: `f${sourceID}`,
    source_id: sourceID,
    type: 'file' as const,
    name,
    path_text: `/${name}`,
    parent_id: 0,
    modified_at: 1700000000,
    created_at: 1700000000,
    size: 1024,
    sha1: `hash-${sourceID}`,
    in_trash: false,
    is_deleted: false,
    highlighted_name: '',
  }
}

describe('SearchPage', () => {
  let wrapper: ReturnType<typeof createTestProvider>

  beforeEach(() => {
    wrapper = createTestProvider()
    createPublicSearchClientSpy.mockReset()
    createPublicSearchClientSpy.mockReturnValue(makeInstantSearchClient([], 0))
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('shows initial state on load', () => {
    setTestURL('/')
    render(<SearchPage />, { wrapper })
    expect(screen.getByText('等待探索')).toBeInTheDocument()
  })

  it('keeps hero mode at viewport height without vertical scrolling', () => {
    setTestURL('/')
    const { container } = render(<SearchPage />, { wrapper })
    const root = container.firstElementChild as HTMLElement | null

    expect(root).not.toBeNull()
    expect(root).toHaveClass('mode-hero')
  })

  it('uses an opaque background for sticky header in docked mode', async () => {
    setTestURL('/')
    const { container } = render(<SearchPage />, { wrapper })
    const user = userEvent.setup()

    await user.type(screen.getByRole('searchbox'), 'test')

    await waitFor(() => {
      const root = container.firstElementChild as HTMLElement | null
      expect(root).not.toBeNull()
      expect(root).toHaveClass('mode-docked')
    })

    const header = container.querySelector('.search-stage') as HTMLElement | null
    expect(header).not.toBeNull()
    expect(header).toHaveClass('search-stage-opaque')
  })

  it('bootstraps public search client when GetSearchConfig enables instantsearch', async () => {
    setTestURL('/')
    let appSearchCalls = 0

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
      http.post('/npan.v1.AppService/AppSearch', () => {
        appSearchCalls += 1
        const response = makeSearchResponse(3, 3)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(createPublicSearchClientSpy).toHaveBeenCalledWith({
        host: 'https://search.example.com',
        indexName: 'npan-public',
        searchApiKey: 'public-search-key',
      })
    })

    expect(appSearchCalls).toBe(0)
  })

  it('uses InstantSearch hits instead of AppSearch results when public search is enabled', async () => {
    setTestURL('/')
    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
    )
    createPublicSearchClientSpy.mockReturnValue(
      makeInstantSearchClient(
        [
          {
            objectID: 'file_1',
            doc_id: 'file_1',
            source_id: 1,
            type: 'file',
            name: 'report.pdf',
            path_text: '/docs/report.pdf',
            parent_id: 0,
            modified_at: 1700000000,
            created_at: 1700000000,
            size: 1024,
            sha1: 'abc',
            in_trash: false,
            is_deleted: false,
            _highlightResult: {
              name: {
                value: '<mark>report</mark>.pdf',
              },
            },
          },
          {
            objectID: 'file_2',
            doc_id: 'file_2',
            source_id: 2,
            type: 'file',
            name: 'manual.pdf',
            path_text: '/docs/manual.pdf',
            parent_id: 0,
            modified_at: 1700000001,
            created_at: 1700000001,
            size: 2048,
            sha1: 'def',
            in_trash: false,
            is_deleted: false,
          },
        ],
        2,
      ),
    )
    render(<SearchPage />, { wrapper })

    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'report{Enter}')

    await waitFor(() => {
      expect(screen.getByTitle('report.pdf')).toBeInTheDocument()
      expect(screen.getByText('已加载 2 / 2 个文件')).toBeInTheDocument()
    })
  })

  it('falls back to legacy AppSearch when GetSearchConfig disables instantsearch', async () => {
    setTestURL('/')
    let appSearchCalls = 0

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse({ instantsearchEnabled: false }))
      }),
      http.post('/npan.v1.AppService/AppSearch', () => {
        appSearchCalls += 1
        const response = makeSearchResponse(3, 3)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    const input = screen.getByRole('searchbox')
    await user.type(input, 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('file1.pdf')).toBeInTheDocument()
      expect(appSearchCalls).toBeGreaterThan(0)
    })

    expect(createPublicSearchClientSpy).not.toHaveBeenCalled()
  })

  it('falls back to legacy AppSearch when public search config is missing required fields', async () => {
    setTestURL('/')
    let appSearchCalls = 0

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(
          makeSearchConfigResponse({
            host: '',
            indexName: '',
            searchApiKey: '',
            instantsearchEnabled: true,
          }),
        )
      }),
      http.post('/npan.v1.AppService/AppSearch', () => {
        appSearchCalls += 1
        const response = makeSearchResponse(1, 1)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('file1.pdf')).toBeInTheDocument()
      expect(appSearchCalls).toBeGreaterThan(0)
    })

    expect(createPublicSearchClientSpy).not.toHaveBeenCalled()
  })

  it('shows results after search', async () => {
    setTestURL('/')
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        const response = makeSearchResponse(3, 3)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    const input = screen.getByRole('searchbox')
    await user.type(input, 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('file1.pdf')).toBeInTheDocument()
      expect(screen.getByText('file2.pdf')).toBeInTheDocument()
      expect(screen.getByText('file3.pdf')).toBeInTheDocument()
    })
  })

  it('refetches on foreground and avoids duplicate foreground refetch bursts', async () => {
    setTestURL('/')
    let requestCount = 0
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        requestCount += 1
        const response = makeSearchResponse(2, 2)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('file1.pdf')).toBeInTheDocument()
      expect(requestCount).toBe(1)
    })

    document.dispatchEvent(new Event('visibilitychange'))
    window.dispatchEvent(new Event('focus'))

    await waitFor(() => {
      expect(requestCount).toBe(2)
    })
  })

  it('shows no results state', async () => {
    setTestURL('/')
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(toConnectSearchResponse([], 0))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'nonexistent{Enter}')

    await waitFor(() => {
      // Empty state card has the description text
      expect(screen.getByText(/没有找到匹配的内容/)).toBeInTheDocument()
    })
  })

  it('shows error state on API failure', async () => {
    setTestURL('/')
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(
          { code: 'internal', message: 'Server error' },
          { status: 500 },
        )
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('加载出错了')).toBeInTheDocument()
    })
  })

  it('returns to initial state on clear', async () => {
    setTestURL('/')
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        const response = makeSearchResponse(1, 1)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('file1.pdf')).toBeInTheDocument()
    })

    // Click clear button
    await user.click(screen.getByLabelText('清空搜索'))

    await waitFor(() => {
      expect(screen.getByText('等待探索')).toBeInTheDocument()
    })
  })

  it('shows result count', async () => {
    setTestURL('/')
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        const response = makeSearchResponse(3, 50)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      // Counter shows "3 / 50"
      expect(screen.getByText('3 / 50')).toBeInTheDocument()
    })
  })

  it('defaults to all filter when file_category missing', async () => {
    setTestURL('/')
    const items = [
      makeItem(1, 'spec.pdf'),
      makeItem(2, 'photo.jpg'),
      makeItem(3, 'movie.mp4'),
    ]
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(toConnectSearchResponse(items, items.length))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByRole('radio', { name: '全部' })).toBeChecked()
      expect(screen.getByText('spec.pdf')).toBeInTheDocument()
      expect(screen.getByText('photo.jpg')).toBeInTheDocument()
      expect(screen.getByText('movie.mp4')).toBeInTheDocument()
    })
  })

  it('restores public search query and file_category from url on first render', async () => {
    setTestURL('/?query=report&page=2&file_category=doc')
    const searchRequests: Array<Array<{ params?: Record<string, unknown> }>> = []

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
    )
    createPublicSearchClientSpy.mockReturnValue(
      makeInstantSearchClient(
        [
          {
            objectID: 'file_1',
            doc_id: 'file_1',
            source_id: 1,
            type: 'file',
            name: 'report.pdf',
            path_text: '/docs/report.pdf',
            parent_id: 0,
            modified_at: 1700000000,
            created_at: 1700000000,
            size: 1024,
            sha1: 'abc',
            in_trash: false,
            is_deleted: false,
            file_category: 'doc',
          },
        ],
        1,
        {
          onSearch: (requests) => {
            searchRequests.push(requests)
          },
        },
      ),
    )

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(screen.getByRole('searchbox')).toHaveValue('report')
    })

    await waitFor(() => {
      expect(screen.getByRole('radio', { name: '文档' })).toBeChecked()
      expect(screen.getByText('report.pdf')).toBeInTheDocument()
    })

    expect(searchRequests.length).toBeGreaterThan(0)
    expect(searchRequests.some((requests) => {
      const params = requests[0]?.params
      return params?.query === 'report'
        && Array.isArray(params?.facetFilters)
        && JSON.stringify(params.facetFilters).includes('file_category:doc')
    })).toBe(true)
  })

  it('falls back to all filter when file_category url value is invalid in public mode', async () => {
    setTestURL('/?file_category=invalid')
    const searchRequests: Array<Array<{ params?: Record<string, unknown> }>> = []

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
    )
    createPublicSearchClientSpy.mockReturnValue(
      makeInstantSearchClient(
        [
          {
            objectID: 'file_1',
            doc_id: 'file_1',
            source_id: 1,
            type: 'file',
            name: 'report.pdf',
            path_text: '/docs/report.pdf',
            parent_id: 0,
            modified_at: 1700000000,
            created_at: 1700000000,
            size: 1024,
            sha1: 'abc',
            in_trash: false,
            is_deleted: false,
            file_category: 'doc',
          },
        ],
        1,
        {
          onSearch: (requests) => {
            searchRequests.push(requests)
          },
        },
      ),
    )

    render(<SearchPage />, { wrapper })

    expect(screen.getByRole('radio', { name: '全部' })).toBeChecked()

    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'report{Enter}')

    await waitFor(() => {
      expect(screen.getByText('report.pdf')).toBeInTheDocument()
    })

    await user.click(screen.getByRole('radio', { name: '图片' }))

    await waitFor(() => {
      const params = new URL(window.location.href).searchParams
      expect(screen.getByRole('radio', { name: '图片' })).toBeChecked()
      expect(params.get('file_category')).toBe('image')
      expect(params.get('ext')).toBeNull()
    })

    expect(searchRequests.some((requests) => {
      const params = requests[0]?.params
      return Array.isArray(params?.facetFilters)
        && JSON.stringify(params.facetFilters).includes('file_category:image')
    })).toBe(true)
  })

  it('updates public search url and uses file_category refinement instead of local filtering', async () => {
    setTestURL('/?query=report')
    const searchRequests: Array<Array<{ params?: Record<string, unknown> }>> = []
    const docHit = {
      objectID: 'file_1',
      doc_id: 'file_1',
      source_id: 1,
      type: 'file',
      name: 'report.pdf',
      path_text: '/docs/report.pdf',
      parent_id: 0,
      modified_at: 1700000000,
      created_at: 1700000000,
      size: 1024,
      sha1: 'abc',
      in_trash: false,
      is_deleted: false,
      file_category: 'doc',
    }
    const imageHit = {
      objectID: 'file_2',
      doc_id: 'file_2',
      source_id: 2,
      type: 'file',
      name: 'manual.pdf',
      path_text: '/docs/manual.pdf',
      parent_id: 0,
      modified_at: 1700000001,
      created_at: 1700000001,
      size: 2048,
      sha1: 'def',
      in_trash: false,
      is_deleted: false,
      file_category: 'image',
    }

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
    )
    createPublicSearchClientSpy.mockReturnValue(
      makeInstantSearchClient(
        [docHit, imageHit],
        2,
        {
          onSearch: (requests) => {
            searchRequests.push(requests)
          },
          resolveResult: (requests) => {
            const facetFilters = requests[0]?.params?.facetFilters
            if (Array.isArray(facetFilters) && JSON.stringify(facetFilters).includes('file_category:doc')) {
              return { hits: [docHit], total: 1 }
            }
            return { hits: [docHit, imageHit], total: 2 }
          },
        },
      ),
    )

    render(<SearchPage />, { wrapper })

    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByRole('searchbox')).toHaveValue('report')
      expect(screen.getByText('report.pdf')).toBeInTheDocument()
      expect(screen.getByText('manual.pdf')).toBeInTheDocument()
    })

    await user.click(screen.getByRole('radio', { name: '文档' }))

    await waitFor(() => {
      expect(screen.getByRole('radio', { name: '文档' })).toBeChecked()
      expect(new URL(window.location.href).searchParams.get('file_category')).toBe('doc')
      expect(screen.getByText('已加载 1 / 1 个文件')).toBeInTheDocument()
      expect(screen.getByText('report.pdf')).toBeInTheDocument()
      expect(screen.queryByText('manual.pdf')).not.toBeInTheDocument()
    })

    expect(searchRequests.some((requests) => {
      const params = requests[0]?.params
      return params?.query === 'report'
        && Array.isArray(params?.facetFilters)
        && JSON.stringify(params.facetFilters).includes('file_category:doc')
    })).toBe(true)
  })

  it('clears public search query and file_category url params on clear', async () => {
    setTestURL('/?query=report')
    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
    )
    createPublicSearchClientSpy.mockReturnValue(
      makeInstantSearchClient(
        [
          {
            objectID: 'file_1',
            doc_id: 'file_1',
            source_id: 1,
            type: 'file',
            name: 'report.pdf',
            path_text: '/docs/report.pdf',
            parent_id: 0,
            modified_at: 1700000000,
            created_at: 1700000000,
            size: 1024,
            sha1: 'abc',
            in_trash: false,
            is_deleted: false,
            file_category: 'doc',
          },
        ],
        1,
      ),
    )

    render(<SearchPage />, { wrapper })

    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByRole('searchbox')).toHaveValue('report')
      expect(screen.getByLabelText('清空搜索')).toBeInTheDocument()
    })

    await user.click(screen.getByRole('radio', { name: '文档' }))

    await waitFor(() => {
      const params = new URL(window.location.href).searchParams
      expect(params.get('file_category')).toBe('doc')
    })

    await user.click(screen.getByLabelText('清空搜索'))

    await waitFor(() => {
      const params = new URL(window.location.href).searchParams
      expect(screen.getByText('等待探索')).toBeInTheDocument()
      expect(params.get('query')).toBeNull()
      expect(params.get('file_category')).toBeNull()
    })
  })

  it('downloads public InstantSearch results via AppDownloadURL instead of hit payload urls', async () => {
    setTestURL('/?query=report')
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null)
    const downloadBodies: unknown[] = []

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
      http.post('/npan.v1.AppService/AppDownloadURL', async ({ request }) => {
        const body: unknown = await request.json()
        downloadBodies.push(body)
        return HttpResponse.json({
          result: {
            fileId: '1',
            downloadUrl: 'https://example.com/rpc-download.pdf',
          },
        })
      }),
    )
    createPublicSearchClientSpy.mockReturnValue(
      makeInstantSearchClient(
        [
          {
            objectID: 'file_1',
            doc_id: 'file_1',
            source_id: 1,
            type: 'file',
            name: 'report.pdf',
            path_text: '/docs/report.pdf',
            parent_id: 0,
            modified_at: 1700000000,
            created_at: 1700000000,
            size: 1024,
            sha1: 'abc',
            in_trash: false,
            is_deleted: false,
            file_category: 'doc',
            downloadUrl: 'https://example.com/hit-download.pdf',
          },
        ],
        1,
      ),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText('report.pdf')).toBeInTheDocument()
    })

    const resultCard = screen.getByText('report.pdf').closest('article')
    expect(resultCard).not.toBeNull()
    if (!resultCard) {
      throw new Error('expected report result card')
    }

    await user.click(within(resultCard).getByRole('button', { name: '下载' }))

    await waitFor(() => {
      expect(downloadBodies).toHaveLength(1)
      expect(openSpy).toHaveBeenCalledWith(
        'https://example.com/rpc-download.pdf',
        '_blank',
        'noopener,noreferrer',
      )
    })

    expect(downloadBodies[0]).toEqual({ fileId: '1' })
    expect(openSpy).not.toHaveBeenCalledWith(
      'https://example.com/hit-download.pdf',
      '_blank',
      'noopener,noreferrer',
    )
  })

  it('keeps download flow available after falling back to legacy AppSearch', async () => {
    setTestURL('/')
    let appSearchCalls = 0
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null)

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse({ instantsearchEnabled: false }))
      }),
      http.post('/npan.v1.AppService/AppSearch', () => {
        appSearchCalls += 1
        return HttpResponse.json(
          toConnectSearchResponse([makeItem(1, 'fallback.pdf')], 1),
        )
      }),
      http.post('/npan.v1.AppService/AppDownloadURL', () => {
        return HttpResponse.json({
          result: {
            fileId: '1',
            downloadUrl: 'https://example.com/fallback.pdf',
          },
        })
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'fallback{Enter}')

    await waitFor(() => {
      expect(screen.getByText('fallback.pdf')).toBeInTheDocument()
      expect(appSearchCalls).toBeGreaterThan(0)
    })

    await user.click(screen.getByRole('button', { name: '下载' }))

    await waitFor(() => {
      expect(openSpy).toHaveBeenCalledWith(
        'https://example.com/fallback.pdf',
        '_blank',
        'noopener,noreferrer',
      )
    })

    expect(createPublicSearchClientSpy).not.toHaveBeenCalled()
  })
})
