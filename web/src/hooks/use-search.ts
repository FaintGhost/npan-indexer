import { useState, useRef, useCallback, useEffect } from 'react'
import { ConnectError } from '@connectrpc/connect'
import { useMutation } from '@connectrpc/connect-query'
import { appSearch as appSearchMethod } from '@/gen/npan/v1/api-AppService_connectquery'
import { fromProtoAppSearchResponse } from '@/lib/connect-app-adapter'
import type { IndexDocument } from '@/lib/schemas'

const DEBOUNCE_MS = 280
const PAGE_SIZE = 30

export function useSearch() {
  const [query, setQueryState] = useState('')
  const [items, setItems] = useState<IndexDocument[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const seqRef = useRef(0)
  const searchMutation = useMutation(appSearchMethod, {
    retry: false,
  })

  const hasMore = items.length < total

  const doSearch = useCallback(async (q: string, p: number, append: boolean) => {
    if (!q.trim()) return

    const seq = ++seqRef.current
    setLoading(true)
    setError(null)

    try {
      const response = await searchMutation.mutateAsync({
        query: q,
        page: BigInt(p),
        pageSize: BigInt(PAGE_SIZE),
      })
      const result = fromProtoAppSearchResponse(response)

      // Ignore stale responses
      if (seq !== seqRef.current) return

      if (append) {
        setItems((prev) => {
          const seen = new Set(prev.map((item) => item.source_id))
          const newItems = result.items.filter((item) => !seen.has(item.source_id))
          return [...prev, ...newItems]
        })
      } else {
        setItems(result.items)
      }
      setTotal(result.total)
      setPage(p)
    } catch (err) {
      if (seq !== seqRef.current) return
      if (err instanceof ConnectError) {
        setError(err.rawMessage || err.message)
      } else {
        setError(err instanceof Error ? err.message : 'Unknown error')
      }
    } finally {
      if (seq === seqRef.current) {
        setLoading(false)
      }
    }
  }, [searchMutation])

  const setQuery = useCallback((q: string) => {
    setQueryState(q)
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }
    if (!q.trim()) {
      setItems([])
      setTotal(0)
      setPage(1)
      return
    }
    debounceRef.current = setTimeout(() => {
      doSearch(q, 1, false)
    }, DEBOUNCE_MS)
  }, [doSearch])

  const searchImmediate = useCallback((q: string) => {
    setQueryState(q)
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }
    doSearch(q, 1, false)
  }, [doSearch])

  const loadMore = useCallback(() => {
    if (loading || !hasMore || !query.trim()) return
    doSearch(query, page + 1, true)
  }, [loading, hasMore, query, page, doSearch])

  const reset = useCallback(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }
    setQueryState('')
    setItems([])
    setTotal(0)
    setPage(1)
    setLoading(false)
    setError(null)
  }, [])

  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [])

  return {
    query,
    items,
    total,
    loading,
    hasMore,
    error,
    setQuery,
    searchImmediate,
    loadMore,
    reset,
  }
}
