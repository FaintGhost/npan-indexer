import { createLazyFileRoute } from '@tanstack/react-router'
import { useRef, useEffect, useCallback } from 'react'
import { useSearch } from '@/hooks/use-search'
import { useDownload } from '@/hooks/use-download'
import { useViewMode } from '@/hooks/use-view-mode'
import { useHotkey } from '@/hooks/use-hotkey'
import { SearchInput } from '@/components/search-input'
import { FileCard } from '@/components/file-card'
import { InitialState, NoResultsState, ErrorState } from '@/components/empty-state'
import { SkeletonCard } from '@/components/skeleton-card'

export const Route = createLazyFileRoute('/')({
  component: SearchPage,
})

export function SearchPage() {
  const inputRef = useRef<HTMLInputElement>(null)
  const sentinelRef = useRef<HTMLDivElement>(null)

  const search = useSearch()
  const download = useDownload()
  const { isDocked, setDocked } = useViewMode()

  useHotkey('k', () => inputRef.current?.focus())

  const handleChange = useCallback((value: string) => {
    search.setQuery(value)
    if (value.trim()) {
      setDocked(true)
    }
  }, [search.setQuery, setDocked])

  const handleSubmit = useCallback(() => {
    if (search.query.trim()) {
      search.searchImmediate(search.query)
      setDocked(true)
    }
  }, [search.query, search.searchImmediate, setDocked])

  const handleClear = useCallback(() => {
    search.reset()
    setDocked(false)
    inputRef.current?.focus()
  }, [search.reset, setDocked])

  // Infinite scroll observer
  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && search.hasMore && !search.loading) {
          search.loadMore()
        }
      },
      { root: null, rootMargin: '180px 0px', threshold: 0.01 },
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [search.hasMore, search.loading, search.loadMore])

  const showInitial = !search.query && search.items.length === 0 && !search.loading
  const showNoResults = !!search.query && !search.loading && search.items.length === 0 && !search.error
  const showError = !!search.error
  const showResults = search.items.length > 0
  const showSkeleton = search.loading && search.items.length === 0

  // Status text
  let statusText = '随时准备为您检索文件'
  let statusError = false
  if (search.loading && search.items.length === 0) {
    statusText = '检索中...'
  } else if (search.loading && search.items.length > 0) {
    statusText = '正在加载更多...'
  } else if (showResults) {
    statusText = `已加载 ${search.items.length} / ${search.total} 个文件`
  } else if (showNoResults) {
    statusText = '未找到相关文件'
  } else if (showError) {
    statusText = search.error ?? '请求失败'
    statusError = true
  }

  return (
    <div className={isDocked ? 'mode-docked' : 'mode-hero'}>
      {/* Search header */}
      <header className="search-stage">
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
                value={search.query}
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
          </div>
        </div>
      </header>

      {/* Results */}
      <main className="mx-auto max-w-3xl px-4 pb-16 sm:px-6 lg:px-8">
        <section className="results-wrap mt-2" aria-live="polite" aria-busy={search.loading}>
          {/* Counter bar */}
          <div className="mb-3 flex items-center justify-between">
            <p className="text-sm text-slate-500">结果列表</p>
            <p className="text-sm font-medium text-slate-600">
              {search.items.length} / {search.total}
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

            {showResults && search.items.map((doc) => (
              <FileCard
                key={doc.source_id}
                doc={doc}
                downloadStatus={download.getStatus(doc.source_id)}
                onDownload={() => download.download(doc.source_id)}
              />
            ))}

            {search.loading && search.items.length > 0 && (
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
