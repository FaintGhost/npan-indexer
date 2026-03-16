import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { createTestProvider } from '../tests/test-providers'
import { SearchPage } from '../routes/index.lazy'
import '../app.css'

const { createPublicSearchClientSpy } = vi.hoisted(() => ({
  createPublicSearchClientSpy: vi.fn(),
}))

vi.mock('@/lib/public-search-client', () => ({
  createPublicSearchClient: createPublicSearchClientSpy,
}))

// Helper to create search response
const GET_SEARCH_CONFIG_PATH = '/npan.v1.AppService/GetSearchConfig'
const DEBOUNCE_MS = 280

type InstantSearchRequest = Array<{
  query?: unknown
  params?: Record<string, unknown>
}>

type InstantSearchRequestLog = Array<Array<{ params?: Record<string, unknown> }>>

type SearchEvent = {
  query: string
  at: number
}

const QUERY_NORMALIZATION_CASES = [
  {
    label: '扩展名查询',
    raw: '规格书.pdf',
    expected: 'pdf 规格书',
  },
  {
    label: '版本号查询',
    raw: 'firmware v3.2.1',
    expected: 'firmware 3.2.1',
  },
  {
    label: '多词组合查询',
    raw: 'mx40 spec pdf',
    expected: 'pdf mx40 spec',
  },
]

const PUBLIC_FILTER_CLAUSES = [
  {
    label: 'type=file',
    pattern: /type\s*(?:=|:)\s*["']?file["']?/i,
  },
  {
    label: 'is_deleted=false',
    pattern: /is_deleted\s*(?:=|:)\s*(?:false|0)/i,
  },
  {
    label: 'in_trash=false',
    pattern: /in_trash\s*(?:=|:)\s*(?:false|0)/i,
  },
] as const

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
  provider?: 'meilisearch' | 'typesense'
  host?: string
  indexName?: string
  searchApiKey?: string
  instantsearchEnabled?: boolean
}) {
  return {
    provider: 'meilisearch',
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
    onSearch?: (requests: InstantSearchRequestLog[number]) => void
    resolveResult?: (
      requests: InstantSearchRequestLog[number],
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

function makePublicHit(sourceID: number, name: string, fileCategory = 'doc') {
  return {
    objectID: `file_${sourceID}`,
    doc_id: `file_${sourceID}`,
    source_id: sourceID,
    type: 'file',
    name,
    path_text: `/docs/${name}`,
    parent_id: 0,
    modified_at: 1700000000,
    created_at: 1700000000,
    size: 1024,
    sha1: `sha1-${sourceID}`,
    in_trash: false,
    is_deleted: false,
    file_category: fileCategory,
    _highlightResult: {
      name: {
        value: `<mark>${name.replace('.pdf', '')}</mark>.pdf`,
      },
    },
  }
}

function getRequestQuery(requests: InstantSearchRequest): string {
  const request = requests[0]
  if (!request) {
    return ''
  }

  if (typeof request.query === 'string' && request.query.trim() !== '') {
    return request.query
  }

  if (typeof request.params?.query === 'string' && request.params.query.trim() !== '') {
    return request.params.query
  }

  if (typeof request.query === 'string') {
    return request.query
  }

  if (typeof request.params?.query === 'string') {
    return request.params.query
  }

  return ''
}

function createTrackedInstantSearchClient(
  options?: {
    hits?: Array<Record<string, unknown>>
    total?: number
  },
) {
  const events: SearchEvent[] = []
  const hits = options?.hits ?? [makePublicHit(1, 'report.pdf')]
  const total = options?.total ?? hits.length
  let startedAt = Date.now()

  const client = makeInstantSearchClient(hits, total, {
    onSearch: (requests) => {
      events.push({
        query: getRequestQuery(requests),
        at: Date.now() - startedAt,
      })
    },
  })

  return {
    client,
    events,
    resetClock: () => {
      startedAt = Date.now()
    },
  }
}

async function advanceTimers(ms: number) {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(ms)
  })
}

async function settleInstantSearch() {
  await act(async () => {
    for (let index = 0; index < 5; index += 1) {
      await Promise.resolve()
    }
  })
}

async function setSearchboxValue(value: string) {
  const input = screen.getByRole('searchbox')
  await act(async () => {
    fireEvent.change(input, { target: { value } })
  })
  return input
}

function getLatestCapturedQuery(
  searchRequests: InstantSearchRequestLog,
): string | null {
  for (let index = searchRequests.length - 1; index >= 0; index -= 1) {
    const query = searchRequests[index]?.[0]?.params?.query
    if (typeof query === 'string' && query.trim() !== '') {
      return query
    }
  }

  return null
}

function assertLegacyAlignedOutboundQuery(
  raw: string,
  expected: string,
  actual: string | null,
): void {
  if (actual !== expected) {
    throw new Error(
      `expected public outbound query for "${raw}" to be legacy-preprocessed as "${expected}", but current public request forwarded "${actual ?? '<missing>'}" directly`,
    )
  }
}

function getRequestFilterValue(params?: Record<string, unknown>): string | undefined {
  if (!params) {
    return undefined
  }

  const filter = params.filter
  if (typeof filter === 'string') {
    return filter
  }

  const filters = params.filters
  if (typeof filters === 'string') {
    return filters
  }

  return undefined
}

function getMissingPublicFilterClauses(filter: string | undefined): string[] {
  if (!filter) {
    return PUBLIC_FILTER_CLAUSES.map((clause) => clause.label)
  }

  return PUBLIC_FILTER_CLAUSES
    .filter((clause) => !clause.pattern.test(filter))
    .map((clause) => clause.label)
}

function hasPublicBaselineFilter(filter: string | undefined): boolean {
  return getMissingPublicFilterClauses(filter).length === 0
}

function assertPublicBaselineFilter(params?: Record<string, unknown>) {
  const filter = getRequestFilterValue(params)
  const missingClauses = getMissingPublicFilterClauses(filter)
  if (missingClauses.length > 0) {
    throw new Error(
      `当前 public 请求缺少默认过滤基线: ${missingClauses.join(', ')}。收到 filter=${filter ?? '(missing)'}`,
    )
  }
}

function hasFileCategoryRefinement(
  params: Record<string, unknown> | undefined,
  value: string,
): boolean {
  if (!params) {
    return false
  }

  const facetFilters = params.facetFilters
  if (Array.isArray(facetFilters) && JSON.stringify(facetFilters).includes(`file_category:${value}`)) {
    return true
  }

  const filter = getRequestFilterValue(params)
  return typeof filter === 'string'
    && new RegExp(`file_category\\s*(?:=|:)\\s*["']?${value}["']?`, 'i').test(filter)
}

function assertFileCategoryComposesWithPublicBaseline(
  params: Record<string, unknown> | undefined,
  value: string,
) {
  const filter = getRequestFilterValue(params)
  const missingClauses = getMissingPublicFilterClauses(filter)
  const hasRefinement = hasFileCategoryRefinement(params, value)

  if (missingClauses.length > 0 || !hasRefinement) {
    throw new Error(
      `当前 public 请求未正确叠加默认过滤与 file_category=${value}。缺失默认过滤=${missingClauses.join(', ') || '(none)'}；filter=${filter ?? '(missing)'}；facetFilters=${JSON.stringify(params?.facetFilters ?? null)}`,
    )
  }
}

function findLastSearchParams(
  requestsLog: InstantSearchRequestLog,
  predicate: (params?: Record<string, unknown>) => boolean,
): Record<string, unknown> | undefined {
  for (let index = requestsLog.length - 1; index >= 0; index -= 1) {
    const params = requestsLog[index]?.[0]?.params
    if (predicate(params)) {
      return params
    }
  }

  return undefined
}

describe('SearchPage', () => {
  let wrapper: ReturnType<typeof createTestProvider>

  beforeEach(() => {
    wrapper = createTestProvider()
    createPublicSearchClientSpy.mockReset()
    createPublicSearchClientSpy.mockReturnValue(makeInstantSearchClient([], 0))
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('shows initial state on load', async () => {
    setTestURL('/')
    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
    )

    render(<SearchPage />, { wrapper })
    expect(screen.getByRole('searchbox')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('Powered by Meilisearch')).toBeInTheDocument()
    })
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
        provider: 'meilisearch',
        host: 'https://search.example.com',
        indexName: 'npan-public',
        searchApiKey: 'public-search-key',
      })
    }, { timeout: 5_000 })

    expect(screen.getByText('Powered by Meilisearch')).toBeInTheDocument()
    expect(appSearchCalls).toBe(0)
  })

  it('bootstraps typesense public search client when GetSearchConfig enables typesense instantsearch', async () => {
    setTestURL('/')

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => HttpResponse.json(makeSearchConfigResponse({
        provider: 'typesense',
        host: 'https://typesense.example.com',
      }))),
    )

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(createPublicSearchClientSpy).toHaveBeenCalledWith({
        provider: 'typesense',
        host: 'https://typesense.example.com',
        indexName: 'npan-public',
        searchApiKey: 'public-search-key',
      })
    }, { timeout: 5_000 })

    expect(screen.getByText('Powered by Typesense')).toBeInTheDocument()
  })

  it('does not issue a real public search request on initial render when query is empty', async () => {
    setTestURL('/')
    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
    )

    const tracked = createTrackedInstantSearchClient()
    createPublicSearchClientSpy.mockReturnValue(tracked.client)

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(createPublicSearchClientSpy).toHaveBeenCalled()
      expect(screen.getByText('等待探索')).toBeInTheDocument()
    })

    expect(
      tracked.events,
      'public 初始空查询不应触发真实检索请求；当前请求说明网络语义与页面初始态不一致',
    ).toEqual([])
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

    await waitFor(() => {
      expect(createPublicSearchClientSpy).toHaveBeenCalled()
    })

    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'report{Enter}')

    await waitFor(() => {
      expect(screen.getByTitle('report.pdf')).toBeInTheDocument()
      expect(screen.getByText('已加载 2 / 2 个文件')).toBeInTheDocument()
    })

    expect(screen.queryByText('结果列表')).not.toBeInTheDocument()
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
    setTestURL('/?query=test')
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
            name: 'file1.pdf',
            path_text: '/docs/file1.pdf',
            parent_id: 0,
            modified_at: 1700000000,
            created_at: 1700000000,
            size: 1024,
            sha1: 'abc',
            in_trash: false,
            is_deleted: false,
          },
          {
            objectID: 'file_2',
            doc_id: 'file_2',
            source_id: 2,
            type: 'file',
            name: 'file2.pdf',
            path_text: '/docs/file2.pdf',
            parent_id: 0,
            modified_at: 1700000001,
            created_at: 1700000001,
            size: 2048,
            sha1: 'def',
            in_trash: false,
            is_deleted: false,
          },
          {
            objectID: 'file_3',
            doc_id: 'file_3',
            source_id: 3,
            type: 'file',
            name: 'file3.pdf',
            path_text: '/docs/file3.pdf',
            parent_id: 0,
            modified_at: 1700000002,
            created_at: 1700000002,
            size: 4096,
            sha1: 'ghi',
            in_trash: false,
            is_deleted: false,
          },
        ],
        50,
      ),
    )

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(screen.getByText('已加载 3 / 50 个文件')).toBeInTheDocument()
    })

    expect(screen.queryByText('结果列表')).not.toBeInTheDocument()
  })

  it('keeps trailing spaces in the public input while typing between words', async () => {
    setTestURL('/')
    const searchRequests: InstantSearchRequestLog = []

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
            name: 'mx40-guide.pdf',
            path_text: '/docs/mx40-guide.pdf',
            parent_id: 0,
            modified_at: 1700000000,
            created_at: 1700000000,
            size: 1024,
            sha1: 'abc',
            in_trash: false,
            is_deleted: false,
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
      expect(createPublicSearchClientSpy).toHaveBeenCalled()
    })

    const user = userEvent.setup()
    const input = screen.getByRole('searchbox')

    await user.type(input, 'mx40 ')
    expect(input).toHaveValue('mx40 ')

    await new Promise((resolve) => setTimeout(resolve, 400))

    await waitFor(() => {
      expect(input).toHaveValue('mx40 ')
      expect(getLatestCapturedQuery(searchRequests)).toBe('mx40')
    })

    await user.type(input, 'pdf')
    expect(input).toHaveValue('mx40 pdf')
  })

  it('uses the compact mobile-first search layout classes', async () => {
    setTestURL('/')
    const { container } = render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(screen.getByText('Powered by Meilisearch')).toBeInTheDocument()
    })

    const searchCard = container.querySelector('.search-card') as HTMLElement | null
    expect(searchCard).not.toBeNull()
    expect(searchCard).toHaveClass('p-4', 'sm:p-7')
    expect(screen.getByRole('button', { name: '搜索' })).toHaveClass('w-full', 'sm:w-auto')
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

    await waitFor(() => {
      expect(createPublicSearchClientSpy).toHaveBeenCalled()
      expect(screen.getByRole('radio', { name: '全部' })).toBeChecked()
    })

    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'report{Enter}')

    await waitFor(() => {
      expect(screen.getByTitle('report.pdf')).toBeInTheDocument()
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
    const tracked = createTrackedInstantSearchClient({
      hits: [makePublicHit(1, 'report.pdf')],
      total: 1,
    })
    createPublicSearchClientSpy.mockReturnValue(tracked.client)

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

    tracked.events.length = 0
    tracked.resetClock()

    await user.click(screen.getByLabelText('清空搜索'))

    await waitFor(() => {
      const params = new URL(window.location.href).searchParams
      expect(screen.getByText('等待探索')).toBeInTheDocument()
      expect(params.get('query')).toBeNull()
      expect(params.get('file_category')).toBeNull()
    })

    expect(
      tracked.events,
      'clear 后回到初始态时不应继续发空查询请求；当前请求说明 clear 只清 UI，未清网络语义',
    ).toEqual([])
  })

  it('auto-searches in public mode after input settles for about 280ms', async () => {
    setTestURL('/')
    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => HttpResponse.json(makeSearchConfigResponse())),
    )

    const tracked = createTrackedInstantSearchClient({
      hits: [makePublicHit(1, 'report.pdf')],
      total: 1,
    })
    createPublicSearchClientSpy.mockReturnValue(tracked.client)

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(createPublicSearchClientSpy).toHaveBeenCalled()
    })

    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-03-08T00:00:00.000Z'))

    tracked.events.length = 0
    tracked.resetClock()

    await setSearchboxValue('report')
    await settleInstantSearch()

    expect(tracked.events).toEqual([])

    await advanceTimers(DEBOUNCE_MS - 1)
    await settleInstantSearch()
    expect(tracked.events).toEqual([])

    vi.setSystemTime(new Date('2026-03-08T00:00:00.280Z'))
    await advanceTimers(1)
    await settleInstantSearch()

    expect(
      tracked.events,
      'public 搜索应在输入停止约 280ms 后自动触发；当前 0 次请求表明仍是 submit-only',
    ).toHaveLength(1)
    expect(tracked.events[0]).toMatchObject({ query: 'report' })
    expect(tracked.events[0]?.at).toBeGreaterThanOrEqual(DEBOUNCE_MS)
    expect(tracked.events[0]?.at).toBeLessThanOrEqual(DEBOUNCE_MS + 1)
    expect(screen.getByTitle('report.pdf')).toBeInTheDocument()
    expect(screen.getByText('已加载 1 / 1 个文件')).toBeInTheDocument()

    await advanceTimers(400)
    await settleInstantSearch()
    expect(new URL(window.location.href).searchParams.get('query')).toBe('report')
  })

  it('submits current public query immediately on Enter before debounce expires', async () => {
    setTestURL('/')
    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => HttpResponse.json(makeSearchConfigResponse())),
    )

    const tracked = createTrackedInstantSearchClient({
      hits: [makePublicHit(1, 'report.pdf')],
      total: 1,
    })
    createPublicSearchClientSpy.mockReturnValue(tracked.client)

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(createPublicSearchClientSpy).toHaveBeenCalled()
    })

    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-03-08T00:00:00.000Z'))

    tracked.events.length = 0
    tracked.resetClock()

    const input = await setSearchboxValue('report')
    await settleInstantSearch()
    expect(tracked.events).toEqual([])

    vi.setSystemTime(new Date('2026-03-08T00:00:00.050Z'))
    await act(async () => {
      fireEvent.keyDown(input, { key: 'Enter', code: 'Enter', charCode: 13 })
    })
    await settleInstantSearch()

    expect(
      tracked.events,
      '按 Enter 应在 debounce 到期前立即触发当前查询；当前 0 次请求表明 public 仍是 submit-only',
    ).toHaveLength(1)
    expect(tracked.events[0]).toMatchObject({ query: 'report', at: 50 })
    expect(screen.getByTitle('report.pdf')).toBeInTheDocument()

    await advanceTimers(DEBOUNCE_MS)
    await settleInstantSearch()
    expect(tracked.events).toHaveLength(1)
  })

  it('submits current public query immediately on search button click before debounce expires', async () => {
    setTestURL('/')
    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => HttpResponse.json(makeSearchConfigResponse())),
    )

    const tracked = createTrackedInstantSearchClient({
      hits: [makePublicHit(1, 'report.pdf')],
      total: 1,
    })
    createPublicSearchClientSpy.mockReturnValue(tracked.client)

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(createPublicSearchClientSpy).toHaveBeenCalled()
    })

    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-03-08T00:00:00.000Z'))

    tracked.events.length = 0
    tracked.resetClock()

    await setSearchboxValue('report')
    await settleInstantSearch()
    expect(tracked.events).toEqual([])

    vi.setSystemTime(new Date('2026-03-08T00:00:00.040Z'))
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: '搜索' }))
    })
    await settleInstantSearch()

    expect(
      tracked.events,
      '点击搜索按钮应在 debounce 到期前立即触发当前查询；当前 0 次请求表明 public 仍是 submit-only',
    ).toHaveLength(1)
    expect(tracked.events[0]).toMatchObject({ query: 'report', at: 40 })
    expect(screen.getByTitle('report.pdf')).toBeInTheDocument()

    await advanceTimers(DEBOUNCE_MS)
    await settleInstantSearch()
    expect(tracked.events).toHaveLength(1)
  })

  it.each(QUERY_NORMALIZATION_CASES)(
    'keeps raw input and URL while requiring legacy-aligned outbound query for $label',
    async ({ raw, expected }) => {
      setTestURL(`/?query=${encodeURIComponent(raw)}`)
      const searchRequests: InstantSearchRequestLog = []

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
        expect(createPublicSearchClientSpy).toHaveBeenCalled()
        expect(screen.getByRole('searchbox')).toHaveValue(raw)
        expect(screen.getByText('report.pdf')).toBeInTheDocument()
      })

      expect(new URL(window.location.href).searchParams.get('query')).toBe(raw)
      assertLegacyAlignedOutboundQuery(raw, expected, getLatestCapturedQuery(searchRequests))
    },
  )

  it('sends public baseline filters in request layer instead of relying on render-time trimming', async () => {
    setTestURL('/?query=report')
    const searchRequests: InstantSearchRequestLog = []
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
    const folderLeakHit = {
      objectID: 'folder_2',
      doc_id: 'folder_2',
      source_id: 2,
      type: 'folder',
      name: 'report folder',
      path_text: '/docs/report folder',
      parent_id: 0,
      modified_at: 1700000001,
      created_at: 1700000001,
      size: 0,
      sha1: '',
      in_trash: false,
      is_deleted: false,
      file_category: 'other',
    }
    const trashLeakHit = {
      objectID: 'file_3',
      doc_id: 'file_3',
      source_id: 3,
      type: 'file',
      name: 'report-trash.pdf',
      path_text: '/docs/report-trash.pdf',
      parent_id: 0,
      modified_at: 1700000002,
      created_at: 1700000002,
      size: 1024,
      sha1: 'trash',
      in_trash: true,
      is_deleted: false,
      file_category: 'doc',
    }
    const deletedLeakHit = {
      objectID: 'file_4',
      doc_id: 'file_4',
      source_id: 4,
      type: 'file',
      name: 'report-deleted.pdf',
      path_text: '/docs/report-deleted.pdf',
      parent_id: 0,
      modified_at: 1700000003,
      created_at: 1700000003,
      size: 1024,
      sha1: 'deleted',
      in_trash: false,
      is_deleted: true,
      file_category: 'doc',
    }

    server.use(
      http.post(GET_SEARCH_CONFIG_PATH, () => {
        return HttpResponse.json(makeSearchConfigResponse())
      }),
    )
    createPublicSearchClientSpy.mockReturnValue(
      makeInstantSearchClient(
        [docHit, folderLeakHit, trashLeakHit, deletedLeakHit],
        4,
        {
          onSearch: (requests) => {
            searchRequests.push(requests)
          },
        },
      ),
    )

    render(<SearchPage />, { wrapper })

    await waitFor(() => {
      expect(screen.getByText('report.pdf')).toBeInTheDocument()
      expect(screen.getByText('report folder')).toBeInTheDocument()
      expect(screen.getByText('report-trash.pdf')).toBeInTheDocument()
      expect(screen.getByText('report-deleted.pdf')).toBeInTheDocument()
      expect(screen.getByText('已加载 4 / 4 个文件')).toBeInTheDocument()
    })

    const reportRequest = findLastSearchParams(
      searchRequests,
      (params) => params?.query === 'report',
    )

    expect(reportRequest, '应捕获 query=report 的 public 搜索请求').toBeDefined()
    assertPublicBaselineFilter(reportRequest)
    expect(hasPublicBaselineFilter(getRequestFilterValue(reportRequest))).toBe(true)
    expect(screen.getByText('已加载 4 / 4 个文件')).toBeInTheDocument()
  })

  it('composes file_category refinement on top of public baseline filters instead of replacing them', async () => {
    setTestURL('/?query=report')
    const searchRequests: InstantSearchRequestLog = []
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
            const params = requests[0]?.params
            if (hasPublicBaselineFilter(getRequestFilterValue(params)) && hasFileCategoryRefinement(params, 'doc')) {
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
      expect(screen.getByText('已加载 2 / 2 个文件')).toBeInTheDocument()
    })

    await user.click(screen.getByRole('radio', { name: '文档' }))

    await waitFor(() => {
      expect(screen.getByRole('radio', { name: '文档' })).toBeChecked()
      expect(new URL(window.location.href).searchParams.get('file_category')).toBe('doc')
    })

    const refinedRequest = findLastSearchParams(
      searchRequests,
      (params) => params?.query === 'report' && hasFileCategoryRefinement(params, 'doc'),
    )

    expect(refinedRequest, '应捕获携带 file_category=doc 的 public 搜索请求').toBeDefined()
    assertFileCategoryComposesWithPublicBaseline(refinedRequest, 'doc')
    expect(screen.getByText('已加载 1 / 1 个文件')).toBeInTheDocument()
    expect(screen.getByText('report.pdf')).toBeInTheDocument()
    expect(screen.queryByText('manual.pdf')).not.toBeInTheDocument()
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
