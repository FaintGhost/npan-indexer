import { useState, useEffect, useCallback, useRef } from 'react'
import { apiGet, apiPost, ApiError } from '@/lib/api-client'
import { SyncProgressSchema } from '@/lib/sync-schemas'
import type { SyncProgress } from '@/lib/sync-schemas'

const POLL_INTERVAL = 3000

export function useSyncProgress(headers: Record<string, string>) {
  const [progress, setProgress] = useState<SyncProgress | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchProgress = useCallback(async () => {
    try {
      const result = await apiGet(
        '/api/v1/admin/sync/full/progress',
        {},
        SyncProgressSchema,
        { headers },
      )
      setProgress(result as SyncProgress)
      setError(null)
      return result
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError(err instanceof Error ? err.message : 'Unknown error')
      }
      return null
    }
  }, [headers])

  const startPolling = useCallback((status: string) => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }

    if (status === 'running') {
      intervalRef.current = setInterval(() => {
        void fetchProgress().then((result) => {
          if (result && result.status !== 'running') {
            if (intervalRef.current) {
              clearInterval(intervalRef.current)
              intervalRef.current = null
            }
          }
        })
      }, POLL_INTERVAL)
    }
  }, [fetchProgress])

  const startSync = useCallback(async (rootFolderIds: number[]) => {
    setLoading(true)
    setError(null)
    try {
      await apiPost('/api/v1/admin/sync/full', {
        root_folder_ids: rootFolderIds,
      }, { headers })
      const result = await fetchProgress()
      if (result) {
        startPolling(result.status)
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError(err instanceof Error ? err.message : 'Unknown error')
      }
    } finally {
      setLoading(false)
    }
  }, [headers, fetchProgress, startPolling])

  const cancelSync = useCallback(async () => {
    setLoading(true)
    try {
      await apiPost('/api/v1/admin/sync/full/cancel', {}, { headers })
      await fetchProgress()
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError(err instanceof Error ? err.message : 'Unknown error')
      }
    } finally {
      setLoading(false)
    }
  }, [headers, fetchProgress])

  // Initial fetch
  useEffect(() => {
    void fetchProgress().then((result) => {
      if (result) {
        startPolling(result.status)
      }
    })

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return { progress, loading, error, startSync, cancelSync, refetch: fetchProgress }
}
