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

export const Route = createLazyFileRoute('/app/')({
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
      { rootMargin: '200px' },
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [search.hasMore, search.loading, search.loadMore])

  const showInitial = !search.query && search.items.length === 0 && !search.loading
  const showNoResults = search.query && !search.loading && search.items.length === 0 && !search.error
  const showError = !!search.error
  const showResults = search.items.length > 0

  return (
    <div className={isDocked ? 'mode-docked' : 'mode-hero'}>
      {/* Search section */}
      <div className="search-stage">
        <div className="mx-auto max-w-3xl px-4">
          {!isDocked && (
            <div className="mb-6 text-center">
              <h1 className="font-[var(--font-display)] text-4xl font-semibold text-slate-900">
                Npan Search
              </h1>
              <p className="mt-2 text-sm text-slate-500">搜索文件名，直接下载</p>
            </div>
          )}
          <div className="search-card rounded-2xl bg-white p-1 shadow-lg">
            <SearchInput
              ref={inputRef}
              value={search.query}
              onChange={handleChange}
              onSubmit={handleSubmit}
              onClear={handleClear}
            />
          </div>
        </div>
      </div>

      {/* Results section */}
      <div className="results-wrap mx-auto max-w-3xl px-4 py-6" aria-live="polite" aria-busy={search.loading}>
        {showResults && (
          <p className="mb-4 text-sm text-slate-500">
            共 {search.total} 个结果
          </p>
        )}

        {showInitial && <InitialState />}
        {showNoResults && <NoResultsState />}
        {showError && <ErrorState />}

        {showResults && (
          <div className="space-y-3">
            {search.items.map((doc) => (
              <FileCard
                key={doc.source_id}
                doc={doc}
                onDownload={(id) => download.download(id)}
              />
            ))}
          </div>
        )}

        {search.loading && (
          <div className="space-y-3 mt-3">
            {Array.from({ length: 3 }, (_, i) => (
              <SkeletonCard key={i} />
            ))}
          </div>
        )}

        {/* Infinite scroll sentinel */}
        <div ref={sentinelRef} className="h-1" />
      </div>
    </div>
  )
}
