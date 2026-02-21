import { useState, useEffect, useCallback, useRef } from 'react'
import { apiGet, apiPost, apiDelete, ApiError } from '@/lib/api-client'
import { SyncProgressSchema } from '@/lib/sync-schemas'
import type { SyncProgress } from '@/lib/sync-schemas'

const POLL_INTERVAL = 2000

export function useSyncProgress(headers: Record<string, string>) {
  const [progress, setProgress] = useState<SyncProgress | null>(null)
  const [loading, setLoading] = useState(false)
  const [initialLoading, setInitialLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchProgress = useCallback(async () => {
    try {
      const result = await apiGet(
        '/api/v1/admin/sync',
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

  const startSync = useCallback(async (rootFolderIds: number[], mode: string = 'auto') => {
    setLoading(true)
    setError(null)
    try {
      await apiPost('/api/v1/admin/sync', {
        root_folder_ids: rootFolderIds,
        resume_progress: true,
        mode,
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
      await apiDelete('/api/v1/admin/sync', { headers })
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

  // Initial fetch — skip when no auth headers yet.
  const hasAuth = 'X-API-Key' in headers
  useEffect(() => {
    if (!hasAuth) return

    setInitialLoading(true)
    void fetchProgress().then((result) => {
      setInitialLoading(false)
      if (result) {
        startPolling(result.status)
      }
    })

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [hasAuth]) // eslint-disable-line react-hooks/exhaustive-deps

  return { progress, loading, initialLoading, error, startSync, cancelSync, refetch: fetchProgress }
}
