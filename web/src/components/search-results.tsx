import { useEffect, useMemo, useRef } from 'react'
import {
  useInfiniteHits,
  useInstantSearch,
  useSearchBox,
} from 'react-instantsearch'
import { FileCard } from '@/components/file-card'
import {
  ErrorState,
  InitialState,
  NoResultsState,
} from '@/components/empty-state'
import { SkeletonCard } from '@/components/skeleton-card'
import { fromMeiliHit, type MeiliHit } from '@/lib/meili-hit-adapter'

export type DownloadStatus = 'idle' | 'loading' | 'success' | 'error'

interface SearchResultsProps {
  download: {
    getStatus: (fileId: number) => DownloadStatus
    download: (fileId: number) => void
  }
}

export function SearchResults({ download }: SearchResultsProps) {
  const sentinelRef = useRef<HTMLDivElement>(null)
  const { items: rawHits, isLastPage, showMore } = useInfiniteHits<MeiliHit>()
  const { status } = useInstantSearch({ catchError: true })

  const { query } = useSearchBox()
  const items = useMemo(
    () => rawHits.map((hit) => fromMeiliHit(hit)),
    [rawHits],
  )

  const hasQuery = query.trim().length > 0
  const loading = hasQuery && (status === 'loading' || status === 'stalled')
  const hasError = status === 'error'
  const hasMore = !isLastPage

  useEffect(() => {
    if (
      !hasQuery ||
      hasError ||
      loading ||
      !hasMore ||
      typeof IntersectionObserver === 'undefined'
    ) {
      return
    }

    const sentinel = sentinelRef.current
    if (!sentinel) {
      return
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) {
          showMore()
        }
      },
      { root: null, rootMargin: '180px 0px', threshold: 0.01 },
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [hasError, hasMore, hasQuery, loading, showMore])

  const showInitial = !hasQuery
  const showNoResults = hasQuery && !loading && !hasError && items.length === 0
  const showResults = items.length > 0
  const showSkeleton = loading && items.length === 0
  const showMoreLoading = loading && items.length > 0 && hasMore

  return (
    <section
      id="search-results"
      className="results-wrap mt-3"
      aria-live="polite"
      aria-busy={loading}
    >
      <div
        className="thin-scrollbar space-y-3"
        style={{ viewTransitionName: 'results-list' }}
      >
        {showInitial && <InitialState />}
        {showNoResults && <NoResultsState />}
        {hasError && <ErrorState />}

        {showSkeleton && (
          <>
            {Array.from({ length: 5 }, (_, index) => (
              <SkeletonCard key={index} delay={index * 120} />
            ))}
          </>
        )}

        {showResults && items.map((doc) => (
          <FileCard
            key={doc.source_id}
            doc={doc}
            downloadStatus={download.getStatus(doc.source_id)}
            onDownload={() => download.download(doc.source_id)}
          />
        ))}

        {showMoreLoading && (
          <>
            {Array.from({ length: 3 }, (_, index) => (
              <SkeletonCard key={`more-${index}`} delay={index * 120} />
            ))}
          </>
        )}
      </div>

      {showResults && hasMore && (
        <div className="mt-4 flex justify-center">
          <button
            type="button"
            onClick={() => showMore()}
            className="rounded-xl border border-slate-200 bg-white/95 px-4 py-2 text-sm font-medium text-slate-700 hover:border-slate-300"
          >
            加载更多结果
          </button>
        </div>
      )}

      <div ref={sentinelRef} className="h-2" />
    </section>
  )
}
