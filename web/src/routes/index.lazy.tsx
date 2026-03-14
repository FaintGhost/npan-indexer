import { ConnectError } from '@connectrpc/connect'
import { useInfiniteQuery } from '@connectrpc/connect-query'
import { createLazyFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import type { KeyboardEvent, ReactNode, RefObject } from 'react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { InstantSearch, useInstantSearch, useSearchBox } from 'react-instantsearch'
import { history } from 'instantsearch.js/es/lib/routers'
import type { RouterProps } from 'instantsearch.js/es/middlewares/createRouterMiddleware'
import { appSearch as appSearchMethod } from '@/gen/npan/v1/api-AppService_connectquery'
import type { AppSearchResponse } from '@/gen/npan/v1/api_pb'
import {
  ErrorState,
  InitialState,
  NoResultsState,
} from '@/components/empty-state'
import { FileCard } from '@/components/file-card'
import { SearchFilters } from '@/components/search-filters'
import { SearchInput } from '@/components/search-input'
import { SearchResults } from '@/components/search-results'
import { SkeletonCard } from '@/components/skeleton-card'
import { fromProtoAppSearchResponse } from '@/lib/connect-app-adapter'
import {
  DEFAULT_FILTER,
  SEARCH_FILTER_OPTIONS,
  matchesSearchFilter,
  normalizeSearchFilter,
  type SearchFilter,
} from '@/lib/file-category'
import {
  createSearchStateMapping,
  type SearchRouteState,
  type SearchUiState,
} from '@/lib/instantsearch-routing'
import { createPublicSearchClient, type PublicSearchClientConfig } from '@/lib/meili-search-client'
import { wrapPublicSearchClient } from '@/lib/public-search-request-adapter'
import { loadSearchConfig, resolveSearchBootstrapMode } from '@/lib/search-config'
import type { IndexDocument } from '@/lib/schemas'
import { useDownload } from '@/hooks/use-download'
import { useHotkey } from '@/hooks/use-hotkey'
import { useViewMode } from '@/hooks/use-view-mode'

const DEBOUNCE_MS = 280
const PAGE_SIZE = 30n
const FOREGROUND_REFETCH_MIN_INTERVAL_MS = 1500
const STALLED_LOADING_MS = 15000

function readFilterFromURL(): SearchFilter {
  if (typeof window === 'undefined') {
    return DEFAULT_FILTER
  }

  const params = new URLSearchParams(window.location.search)
  return normalizeSearchFilter(params.get('file_category'))
}

function replaceURLParams(updater: (params: URLSearchParams) => void): void {
  if (typeof window === 'undefined') {
    return
  }

  const params = new URLSearchParams(window.location.search)
  updater(params)
  const search = params.toString()
  const nextURL = `${window.location.pathname}${search ? `?${search}` : ''}${window.location.hash}`
  window.history.replaceState({}, '', nextURL)
}

export const Route = createLazyFileRoute('/')({
  component: SearchPage,
})

function toErrorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    return err.rawMessage || err.message
  }
  if (err instanceof Error) {
    return err.message
  }
  return 'Unknown error'
}

function mergePages(pages: AppSearchResponse[]): { items: IndexDocument[]; total: number } {
  const seen = new Set<number>()
  const items: IndexDocument[] = []
  let total = 0

  for (const page of pages) {
    const mapped = fromProtoAppSearchResponse(page)
    if (total === 0) {
      total = mapped.total
    }
    for (const item of mapped.items) {
      if (seen.has(item.source_id)) {
        continue
      }
      seen.add(item.source_id)
      items.push(item)
    }
  }

  return { items, total }
}

type LegacySearchQuery = ReturnType<
  typeof useInfiniteQuery<
    typeof appSearchMethod.input,
    typeof appSearchMethod.output,
    'page'
  >
>

function LegacySearchResults({
  activeQuery,
  activeFilter,
  searchQuery,
  loading,
  error,
  download,
}: {
  activeQuery: string
  activeFilter: SearchFilter
  searchQuery: LegacySearchQuery
  loading: boolean
  error: string | null
  download: ReturnType<typeof useDownload>
}) {
  const sentinelRef = useRef<HTMLDivElement>(null)

  const searchState = useMemo(
    () => mergePages(searchQuery.data?.pages ?? []),
    [searchQuery.data?.pages],
  )
  const items = searchState.items
  const filteredItems = useMemo(
    () => items.filter((item) => matchesSearchFilter(item.name, activeFilter)),
    [activeFilter, items],
  )
  const total = searchState.total
  const hasMore = Boolean(searchQuery.hasNextPage)

  const loadMore = useCallback(() => {
    if (!hasMore || searchQuery.isFetchingNextPage || !activeQuery.trim()) {
      return
    }
    void searchQuery.fetchNextPage()
  }, [activeQuery, hasMore, searchQuery])

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) {
      return
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          loadMore()
        }
      },
      { root: null, rootMargin: '180px 0px', threshold: 0.01 },
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore])

  const showInitial = !activeQuery && filteredItems.length === 0 && !loading
  const showNoResults = !!activeQuery && !loading && filteredItems.length === 0 && !error
  const showError = !!error
  const showResults = filteredItems.length > 0
  const showSkeleton = loading && filteredItems.length === 0

  return (
    <section
      id="search-results"
      className="results-wrap mt-3"
      aria-live="polite"
      aria-busy={loading}
    >
      <div className="frost-panel mb-4 flex items-center justify-between rounded-2xl px-4 py-3">
        <p className="text-sm font-medium text-slate-700">结果列表</p>
        <p className="font-mono text-sm font-semibold text-slate-700">
          {filteredItems.length} / {total}
        </p>
      </div>

      <div className="thin-scrollbar space-y-3" style={{ viewTransitionName: 'results-list' }}>
        {showInitial && <InitialState />}
        {showNoResults && <NoResultsState />}
        {showError && <ErrorState />}

        {showSkeleton && (
          <>
            {Array.from({ length: 5 }, (_, i) => (
              <SkeletonCard key={i} delay={i * 120} />
            ))}
          </>
        )}

        {showResults && filteredItems.map((doc) => (
          <FileCard
            key={doc.source_id}
            doc={doc}
            downloadStatus={download.getStatus(doc.source_id)}
            onDownload={() => download.download(doc.source_id)}
          />
        ))}

        {searchQuery.isFetchingNextPage && filteredItems.length > 0 && (
          <>
            {Array.from({ length: 3 }, (_, i) => (
              <SkeletonCard key={`more-${i}`} delay={i * 120} />
            ))}
          </>
        )}
      </div>

      <div ref={sentinelRef} className="h-2" />
    </section>
  )
}

function SearchPageFrame({
  isDocked,
  statusText,
  statusError,
  inputRef,
  query,
  onChange,
  onSubmit,
  onClear,
  filters,
  results,
}: {
  isDocked: boolean
  statusText: string
  statusError: boolean
  inputRef: RefObject<HTMLInputElement | null>
  query: string
  onChange: (value: string) => void
  onSubmit: () => void
  onClear: () => void
  filters: ReactNode
  results: ReactNode
}) {
  return (
    <div className={`${isDocked ? 'mode-docked' : 'mode-hero'} relative`}>
      <a href="#search-results" className="skip-link">
        跳到结果
      </a>
      <header className={`search-stage${isDocked ? ' search-stage-opaque' : ''}`}>
        <div className="mx-auto w-full max-w-5xl px-4 sm:px-6 lg:px-8">
          <div className="search-card frost-panel w-full rounded-[2rem] p-5 sm:p-7">
            <div className="flex flex-col gap-4 border-b border-slate-200/70 pb-4 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <h1 className="font-display text-4xl font-semibold leading-[0.95] tracking-[-0.03em] text-slate-900 sm:text-[3.3rem]">
                  Npan Search
                </h1>
                <p className="mt-2 max-w-[44ch] text-sm leading-6 text-slate-600">
                  像搜索引擎一样查找文件，命中后直接下载。
                </p>
              </div>
              <div className="inline-flex items-center rounded-xl border border-blue-200 bg-blue-50 px-3 py-1.5 text-xs font-semibold text-blue-800">
                Powered by Local Search
              </div>
            </div>

            <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-[1fr_auto]">
              <SearchInput
                ref={inputRef}
                value={query}
                onChange={onChange}
                onSubmit={onSubmit}
                onClear={onClear}
              />
              <button
                type="button"
                onClick={onSubmit}
                className="action-btn-primary h-12 px-6 text-sm font-semibold"
              >
                搜索
              </button>
            </div>

            <div className="mt-3 flex min-h-5 flex-wrap items-center justify-between gap-2">
              <p className={`text-xs transition-colors duration-300 ${statusError ? 'font-medium text-rose-600' : 'text-slate-600'}`}>
                {statusText}
              </p>
            </div>

            {filters}
          </div>
        </div>
      </header>

      <main className="mx-auto w-full max-w-5xl px-4 pb-20 sm:px-6 lg:px-8">
        {results}
      </main>
    </div>
  )
}

function LegacySearchPage({
  inputRef,
  download,
  isDocked,
  setDocked,
  isBootstrapLoading,
}: {
  inputRef: RefObject<HTMLInputElement | null>
  download: ReturnType<typeof useDownload>
  isDocked: boolean
  setDocked: (docked: boolean) => void
  isBootstrapLoading: boolean
}) {
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const loadingSinceRef = useRef<number | null>(null)
  const lastForegroundRefetchAtRef = useRef(0)

  const [query, setQuery] = useState('')
  const [activeQuery, setActiveQuery] = useState('')
  const [activeFilter, setActiveFilter] = useState<SearchFilter>(() => readFilterFromURL())

  const searchEnabled = activeQuery.trim().length > 0
  const legacySearchEnabled = searchEnabled && !isBootstrapLoading

  const searchQuery = useInfiniteQuery(
    appSearchMethod,
    {
      query: activeQuery,
      page: 1n,
      pageSize: PAGE_SIZE,
    },
    {
      enabled: legacySearchEnabled,
      retry: false,
      refetchOnReconnect: true,
      pageParamKey: 'page',
      getNextPageParam: (lastPage, allPages) => {
        const result = lastPage.result
        if (!result) {
          return undefined
        }

        const loadedCount = allPages.reduce(
          (sum, page) => sum + (page.result?.items.length ?? 0),
          0,
        )
        const total = Number(result.total)
        if (loadedCount >= total) {
          return undefined
        }

        return BigInt(allPages.length + 1)
      },
    },
  )

  const error = searchQuery.error
    ? toErrorMessage(searchQuery.error)
    : null
  const loading = isBootstrapLoading || (legacySearchEnabled && (searchQuery.isPending || searchQuery.isFetching))

  const clearDebounce = useCallback(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
      debounceRef.current = null
    }
  }, [])

  const syncBootstrapQueryParam = useCallback((nextQuery: string) => {
    if (!isBootstrapLoading) {
      return
    }

    const normalizedQuery = nextQuery.trim()
    replaceURLParams((params) => {
      if (normalizedQuery) {
        params.set('query', normalizedQuery)
      } else {
        params.delete('query')
        params.delete('page')
      }
    })
  }, [isBootstrapLoading])

  const queueDebouncedSearch = useCallback((nextQuery: string) => {
    clearDebounce()
    if (!nextQuery.trim()) {
      setActiveQuery('')
      return
    }
    debounceRef.current = setTimeout(() => {
      setActiveQuery(nextQuery)
      debounceRef.current = null
    }, DEBOUNCE_MS)
  }, [clearDebounce])

  const handleChange = useCallback((value: string) => {
    setQuery(value)
    syncBootstrapQueryParam(value)

    if (!value.trim()) {
      clearDebounce()
      setActiveQuery('')
      setDocked(false)
      setActiveFilter(DEFAULT_FILTER)
      replaceURLParams((params) => {
        params.delete('query')
        params.delete('file_category')
      })
      return
    }

    setDocked(true)
    queueDebouncedSearch(value)
  }, [clearDebounce, queueDebouncedSearch, setDocked, syncBootstrapQueryParam])

  const handleSubmit = useCallback(() => {
    if (!query.trim()) {
      return
    }

    clearDebounce()
    syncBootstrapQueryParam(query)
    setDocked(true)

    if (activeQuery === query) {
      void searchQuery.refetch()
      return
    }

    setActiveQuery(query)
  }, [activeQuery, clearDebounce, query, searchQuery, setDocked, syncBootstrapQueryParam])

  const handleClear = useCallback(() => {
    clearDebounce()
    setQuery('')
    setActiveQuery('')
    setDocked(false)
    setActiveFilter(DEFAULT_FILTER)
    replaceURLParams((params) => {
      params.delete('query')
      params.delete('file_category')
    })
    inputRef.current?.focus()
  }, [clearDebounce, inputRef, setDocked])

  const handleFilterChange = useCallback((filter: SearchFilter) => {
    setActiveFilter(filter)
    replaceURLParams((params) => {
      if (filter === DEFAULT_FILTER) {
        params.delete('file_category')
      } else {
        params.set('file_category', filter)
      }
    })
  }, [])

  const handleFilterKeyDown = useCallback((event: KeyboardEvent, current: SearchFilter) => {
    if (!['ArrowRight', 'ArrowDown', 'ArrowLeft', 'ArrowUp'].includes(event.key)) {
      return
    }
    event.preventDefault()
    const currentIndex = SEARCH_FILTER_OPTIONS.findIndex((option) => option.value === current)
    if (currentIndex < 0) {
      return
    }
    const isForward = event.key === 'ArrowRight' || event.key === 'ArrowDown'
    const nextIndex = isForward
      ? (currentIndex + 1) % SEARCH_FILTER_OPTIONS.length
      : (currentIndex - 1 + SEARCH_FILTER_OPTIONS.length) % SEARCH_FILTER_OPTIONS.length
    const nextFilter = SEARCH_FILTER_OPTIONS[nextIndex]?.value
    if (nextFilter) {
      handleFilterChange(nextFilter)
    }
  }, [handleFilterChange])

  const maybeRefetchOnForeground = useCallback(() => {
    if (!activeQuery.trim()) {
      return
    }

    const now = Date.now()
    if (now - lastForegroundRefetchAtRef.current < FOREGROUND_REFETCH_MIN_INTERVAL_MS) {
      return
    }

    const loadingSince = loadingSinceRef.current
    const stalledLoading = loadingSince !== null && now - loadingSince >= STALLED_LOADING_MS
    if ((searchQuery.isPending || searchQuery.isFetching || searchQuery.isFetchingNextPage) && !stalledLoading) {
      return
    }

    lastForegroundRefetchAtRef.current = now
    void searchQuery.refetch()
  }, [activeQuery, searchQuery])

  useEffect(() => {
    return () => {
      clearDebounce()
    }
  }, [clearDebounce])

  useEffect(() => {
    if (loading) {
      if (loadingSinceRef.current === null) {
        loadingSinceRef.current = Date.now()
      }
      return
    }
    loadingSinceRef.current = null
  }, [loading])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    const onPopState = () => {
      setActiveFilter(readFilterFromURL())
    }

    window.addEventListener('popstate', onPopState)
    return () => window.removeEventListener('popstate', onPopState)
  }, [])

  useEffect(() => {
    if (typeof window === 'undefined' || typeof document === 'undefined') {
      return
    }

    const onForeground = () => {
      if (document.visibilityState === 'hidden') {
        return
      }
      maybeRefetchOnForeground()
    }

    document.addEventListener('visibilitychange', onForeground)
    window.addEventListener('focus', onForeground)

    return () => {
      document.removeEventListener('visibilitychange', onForeground)
      window.removeEventListener('focus', onForeground)
    }
  }, [maybeRefetchOnForeground])

  let statusText = '随时准备为您检索文件'
  let statusError = false
  if (loading) {
    statusText = '检索中...'
  } else if (error) {
    statusText = error
    statusError = true
  }

  const legacyFilters = (
    <div className="mt-4" role="radiogroup" aria-label="结果筛选">
      <div className="flex flex-wrap gap-2.5">
        {SEARCH_FILTER_OPTIONS.map((option) => {
          const checked = activeFilter === option.value
          return (
            <button
              key={option.value}
              type="button"
              role="radio"
              aria-checked={checked}
              tabIndex={checked ? 0 : -1}
              onClick={() => handleFilterChange(option.value)}
              onKeyDown={(event) => handleFilterKeyDown(event, option.value)}
              className={checked
                ? 'rounded-xl border border-blue-200 bg-blue-50 px-3 py-1.5 text-xs font-semibold text-blue-800 shadow-sm'
                : 'rounded-xl border border-slate-200 bg-white/95 px-3 py-1.5 text-xs font-medium text-slate-600 hover:border-slate-300'}
            >
              {option.label}
            </button>
          )
        })}
      </div>
    </div>
  )

  const legacyResults = !searchEnabled ? (
    <section
      id="search-results"
      className="results-wrap mt-3"
      aria-live="polite"
      aria-busy={false}
    >
      <div className="frost-panel mb-4 rounded-2xl px-4 py-3">
        <div className="flex items-center justify-between gap-3">
          <p className="text-sm font-medium text-slate-700">结果列表</p>
          <p className="font-mono text-sm font-semibold text-slate-700">0 / 0</p>
        </div>
        <p className="mt-2 text-xs text-slate-600">随时准备为您检索文件</p>
      </div>
      <div className="thin-scrollbar space-y-3" style={{ viewTransitionName: 'results-list' }}>
        <InitialState />
      </div>
    </section>
  ) : (
    <LegacySearchResults
      activeQuery={activeQuery}
      activeFilter={activeFilter}
      searchQuery={searchQuery}
      loading={loading}
      error={error}
      download={download}
    />
  )

  return (
    <SearchPageFrame
      isDocked={isDocked}
      statusText={statusText}
      statusError={statusError}
      inputRef={inputRef}
      query={query}
      onChange={handleChange}
      onSubmit={handleSubmit}
      onClear={handleClear}
      filters={legacyFilters}
      results={legacyResults}
    />
  )
}

function PublicSearchBody({
  inputRef,
  download,
  isDocked,
  setDocked,
}: {
  inputRef: RefObject<HTMLInputElement | null>
  download: ReturnType<typeof useDownload>
  isDocked: boolean
  setDocked: (docked: boolean) => void
}) {
  const { query, refine } = useSearchBox()
  const { status, error, setUiState } = useInstantSearch<SearchUiState>({ catchError: true })
  const [inputValue, setInputValue] = useState(query)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const previousQueryRef = useRef(query)

  useEffect(() => {
    if (query.trim()) {
      setDocked(true)
    }
  }, [query, setDocked])

  useEffect(() => {
    if (query === previousQueryRef.current) {
      return
    }

    previousQueryRef.current = query
    setInputValue(query)
    setDocked(query.trim().length > 0)
  }, [query, setDocked])

  const clearDebounce = useCallback(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
      debounceRef.current = null
    }
  }, [])

  const resetSearchState = useCallback(() => {
    setUiState(() => ({}))
  }, [setUiState])

  const commitQuery = useCallback((nextQuery: string) => {
    const trimmedQuery = nextQuery.trim()
    if (!trimmedQuery) {
      setInputValue('')
      resetSearchState()
      setDocked(false)
      return
    }

    setInputValue(trimmedQuery)
    setDocked(true)
    refine(trimmedQuery)
  }, [refine, resetSearchState, setDocked])

  const handleChange = useCallback((value: string) => {
    setInputValue(value)

    if (!value.trim()) {
      clearDebounce()
      resetSearchState()
      setDocked(false)
      return
    }

    setDocked(true)
    clearDebounce()
    debounceRef.current = setTimeout(() => {
      debounceRef.current = null
      commitQuery(value)
    }, DEBOUNCE_MS)
  }, [clearDebounce, commitQuery, resetSearchState, setDocked])

  const handleSubmit = useCallback(() => {
    clearDebounce()
    commitQuery(inputRef.current?.value ?? inputValue)
  }, [clearDebounce, commitQuery, inputRef, inputValue])

  const handleClear = useCallback(() => {
    clearDebounce()
    setInputValue('')
    resetSearchState()
    setDocked(false)
    inputRef.current?.focus()
  }, [clearDebounce, inputRef, resetSearchState, setDocked])

  useEffect(() => {
    return () => {
      clearDebounce()
    }
  }, [clearDebounce])

  let statusText = '随时准备为您检索文件'
  let statusError = false
  if (query.trim() && (status === 'loading' || status === 'stalled')) {
    statusText = '检索中...'
  } else if (status === 'error') {
    statusText = toErrorMessage(error)
    statusError = true
  }

  return (
    <SearchPageFrame
      isDocked={isDocked}
      statusText={statusText}
      statusError={statusError}
      inputRef={inputRef}
      query={inputValue}
      onChange={handleChange}
      onSubmit={handleSubmit}
      onClear={handleClear}
      filters={<SearchFilters />}
      results={<SearchResults download={download} />}
    />
  )
}

function PublicSearchPage({
  inputRef,
  download,
  isDocked,
  setDocked,
  publicSearch,
  publicSearchConfig,
}: {
  inputRef: RefObject<HTMLInputElement | null>
  download: ReturnType<typeof useDownload>
  isDocked: boolean
  setDocked: (docked: boolean) => void
  publicSearch: ReturnType<typeof createPublicSearchClient>
  publicSearchConfig: PublicSearchClientConfig
}) {
  const routing = useMemo<RouterProps<SearchUiState, SearchRouteState>>(() => ({
    router: history<SearchRouteState>({
      cleanUrlOnDispose: false,
    }),
    stateMapping: createSearchStateMapping(publicSearchConfig.indexName),
  }), [publicSearchConfig.indexName])

  return (
    <InstantSearch<SearchUiState, SearchRouteState>
      indexName={publicSearchConfig.indexName}
      searchClient={publicSearch.searchClient}
      routing={routing}
      future={{ preserveSharedStateOnUnmount: true }}
    >
      <PublicSearchBody
        inputRef={inputRef}
        download={download}
        isDocked={isDocked}
        setDocked={setDocked}
      />
    </InstantSearch>
  )
}

export function SearchPage() {
  const inputRef = useRef<HTMLInputElement>(null)
  const download = useDownload()
  const { isDocked, setDocked } = useViewMode()

  useHotkey('k', () => inputRef.current?.focus())

  const searchConfigQuery = useQuery({
    queryKey: ['search-config'],
    queryFn: () => loadSearchConfig(),
    staleTime: Number.POSITIVE_INFINITY,
    retry: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  })

  const bootstrapMode = useMemo(() => {
    if (searchConfigQuery.isError) {
      return 'legacy' as const
    }
    if (!searchConfigQuery.data) {
      return 'loading' as const
    }
    return resolveSearchBootstrapMode(searchConfigQuery.data)
  }, [searchConfigQuery.data, searchConfigQuery.isError])

  const publicSearchConfig = useMemo<PublicSearchClientConfig | null>(() => {
    if (bootstrapMode !== 'public' || !searchConfigQuery.data) {
      return null
    }

    return {
      host: searchConfigQuery.data.host,
      indexName: searchConfigQuery.data.indexName,
      searchApiKey: searchConfigQuery.data.searchApiKey,
    }
  }, [bootstrapMode, searchConfigQuery.data])

  const publicSearch = useMemo(() => {
    if (!publicSearchConfig) {
      return null
    }

    return wrapPublicSearchClient(createPublicSearchClient(publicSearchConfig))
  }, [publicSearchConfig])

  if (bootstrapMode === 'public' && publicSearch && publicSearchConfig) {
    return (
      <PublicSearchPage
        inputRef={inputRef}
        download={download}
        isDocked={isDocked}
        setDocked={setDocked}
        publicSearch={publicSearch}
        publicSearchConfig={publicSearchConfig}
      />
    )
  }

  return (
    <LegacySearchPage
      inputRef={inputRef}
      download={download}
      isDocked={isDocked}
      setDocked={setDocked}
      isBootstrapLoading={bootstrapMode === 'loading'}
    />
  )
}
