import { ConnectError } from '@connectrpc/connect'
import { useInfiniteQuery } from '@connectrpc/connect-query'
import { createLazyFileRoute } from '@tanstack/react-router'
import { useRef, useEffect, useCallback, useMemo, useState } from 'react'
import { appSearch as appSearchMethod } from '@/gen/npan/v1/api-AppService_connectquery'
import type { AppSearchResponse } from '@/gen/npan/v1/api_pb'
import { useDownload } from '@/hooks/use-download'
import { useViewMode } from '@/hooks/use-view-mode'
import { useHotkey } from '@/hooks/use-hotkey'
import { SearchInput } from '@/components/search-input'
import { FileCard } from '@/components/file-card'
import { InitialState, NoResultsState, ErrorState } from '@/components/empty-state'
import { SkeletonCard } from '@/components/skeleton-card'
import { fromProtoAppSearchResponse } from '@/lib/connect-app-adapter'
import type { IndexDocument } from '@/lib/schemas'
import {
  DEFAULT_FILTER,
  type SearchFilter,
  matchesSearchFilter,
  normalizeSearchFilter,
} from '@/lib/file-category'

const DEBOUNCE_MS = 280
const PAGE_SIZE = 30n
const FILTER_OPTIONS: Array<{ value: SearchFilter; label: string }> = [
  { value: 'all', label: '全部' },
  { value: 'doc', label: '文档' },
  { value: 'image', label: '图片' },
  { value: 'video', label: '视频' },
  { value: 'archive', label: '压缩包' },
  { value: 'other', label: '其他' },
]

function readFilterFromURL(): SearchFilter {
  if (typeof window === 'undefined') {
    return DEFAULT_FILTER
  }
  const params = new URLSearchParams(window.location.search)
  return normalizeSearchFilter(params.get('ext'))
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

export function SearchPage() {
  const inputRef = useRef<HTMLInputElement>(null)
  const sentinelRef = useRef<HTMLDivElement>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const [query, setQuery] = useState('')
  const [activeQuery, setActiveQuery] = useState('')
  const [activeFilter, setActiveFilter] = useState<SearchFilter>(() => readFilterFromURL())

  const download = useDownload()
  const { isDocked, setDocked } = useViewMode()

  useHotkey('k', () => inputRef.current?.focus())

  const searchQuery = useInfiniteQuery(
    appSearchMethod,
    {
      query: activeQuery,
      page: 1n,
      pageSize: PAGE_SIZE,
    },
    {
      enabled: activeQuery.trim().length > 0,
      retry: false,
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
  const error = searchQuery.error ? toErrorMessage(searchQuery.error) : null
  const searchEnabled = activeQuery.trim().length > 0
  const loading = searchEnabled && (searchQuery.isPending || searchQuery.isFetching)
  const hasMore = Boolean(searchQuery.hasNextPage)

  const clearDebounce = useCallback(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
      debounceRef.current = null
    }
  }, [])

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
    if (value.trim()) {
      setDocked(true)
    }
    queueDebouncedSearch(value)
  }, [queueDebouncedSearch, setDocked])

  const handleSubmit = useCallback(() => {
    if (!query.trim()) {
      return
    }

    clearDebounce()
    if (activeQuery === query) {
      void searchQuery.refetch()
    } else {
      setActiveQuery(query)
    }
    setDocked(true)
  }, [activeQuery, clearDebounce, query, searchQuery, setDocked])

  const handleClear = useCallback(() => {
    clearDebounce()
    setQuery('')
    setActiveQuery('')
    setDocked(false)
    setActiveFilter(DEFAULT_FILTER)
    replaceURLParams((params) => {
      params.delete('q')
      params.delete('ext')
    })
    inputRef.current?.focus()
  }, [clearDebounce, setDocked])

  const handleFilterChange = useCallback((filter: SearchFilter) => {
    setActiveFilter(filter)
    replaceURLParams((params) => {
      if (filter === DEFAULT_FILTER) {
        params.delete('ext')
      } else {
        params.set('ext', filter)
      }
    })
  }, [])

  const handleFilterKeyDown = useCallback((event: React.KeyboardEvent, current: SearchFilter) => {
    if (!['ArrowRight', 'ArrowDown', 'ArrowLeft', 'ArrowUp'].includes(event.key)) {
      return
    }
    event.preventDefault()
    const currentIndex = FILTER_OPTIONS.findIndex((option) => option.value === current)
    if (currentIndex < 0) {
      return
    }
    const isForward = event.key === 'ArrowRight' || event.key === 'ArrowDown'
    const nextIndex = isForward
      ? (currentIndex + 1) % FILTER_OPTIONS.length
      : (currentIndex - 1 + FILTER_OPTIONS.length) % FILTER_OPTIONS.length
    const nextFilter = FILTER_OPTIONS[nextIndex]?.value
    if (nextFilter) {
      handleFilterChange(nextFilter)
    }
  }, [handleFilterChange])

  const loadMore = useCallback(() => {
    if (!hasMore || searchQuery.isFetchingNextPage || !activeQuery.trim()) {
      return
    }
    void searchQuery.fetchNextPage()
  }, [activeQuery, hasMore, searchQuery])

  useEffect(() => {
    return () => {
      clearDebounce()
    }
  }, [clearDebounce])

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

  // Infinite scroll observer
  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

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

  const showInitial = !query && filteredItems.length === 0 && !loading
  const showNoResults = !!activeQuery && !loading && filteredItems.length === 0 && !error
  const showError = !!error
  const showResults = filteredItems.length > 0
  const showSkeleton = loading && filteredItems.length === 0

  // Status text
  let statusText = '随时准备为您检索文件'
  let statusError = false
  if (loading && filteredItems.length === 0) {
    statusText = '检索中...'
  } else if (searchQuery.isFetchingNextPage && filteredItems.length > 0) {
    statusText = '正在加载更多...'
  } else if (showResults) {
    statusText = `已加载 ${filteredItems.length} / ${total} 个文件`
  } else if (showNoResults) {
    statusText = '未找到相关文件'
  } else if (showError) {
    statusText = error ?? '请求失败'
    statusError = true
  }

  return (
    <div className={isDocked ? 'mode-docked' : 'mode-hero'}>
      {/* Search header */}
      <header className={`search-stage${isDocked ? ' search-stage-opaque' : ''}`}>
        <div className="mx-auto w-full max-w-3xl px-4 sm:px-6 lg:px-8">
          <div className="search-card w-full rounded-3xl border border-slate-200/90 bg-white p-4 sm:p-6">
            {/* Title row */}
            <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <h1 className="font-display text-4xl font-semibold leading-tight tracking-tight text-slate-900">
                  Npan Search
                </h1>
                <p className="mt-1 text-sm text-slate-500">
                  像搜索引擎一样查找文件，命中后直接下载。
                </p>
              </div>
              <div className="rounded-full border border-blue-100 bg-blue-50 px-3 py-1 text-xs font-medium text-blue-700">
                Powered by Meilisearch
              </div>
            </div>

            {/* Search input + button */}
            <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-[1fr_auto]">
              <SearchInput
                ref={inputRef}
                value={query}
                onChange={handleChange}
                onSubmit={handleSubmit}
                onClear={handleClear}
              />
              <button
                type="button"
                onClick={handleSubmit}
                className="h-12 rounded-xl bg-blue-600 px-5 text-sm font-semibold text-white shadow-md shadow-blue-200 transition hover:bg-blue-500 active:scale-[0.99]"
              >
                搜索
              </button>
            </div>

            {/* Status text */}
            <p className={`mt-3 min-h-5 text-xs transition-colors duration-300 ${statusError ? 'font-medium text-rose-600' : 'text-slate-500'}`}>
              {statusText}
            </p>

            <div className="mt-4" role="radiogroup" aria-label="结果筛选">
              <div className="flex flex-wrap gap-2">
                {FILTER_OPTIONS.map((option) => {
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
                        ? 'rounded-full border border-blue-300 bg-blue-50 px-3 py-1.5 text-xs font-medium text-blue-700'
                        : 'rounded-full border border-slate-200 bg-white px-3 py-1.5 text-xs font-medium text-slate-600 hover:border-slate-300'}
                    >
                      {option.label}
                    </button>
                  )
                })}
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Results */}
      <main className="mx-auto max-w-3xl px-4 pb-16 sm:px-6 lg:px-8">
        <section className="results-wrap mt-2" aria-live="polite" aria-busy={loading}>
          {/* Counter bar */}
          <div className="mb-3 flex items-center justify-between">
            <p className="text-sm text-slate-500">结果列表</p>
            <p className="text-sm font-medium text-slate-600">
              {filteredItems.length} / {total}
            </p>
          </div>

          {/* List */}
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

          {/* Infinite scroll sentinel */}
          <div ref={sentinelRef} className="h-2" />
        </section>
      </main>
    </div>
  )
}
