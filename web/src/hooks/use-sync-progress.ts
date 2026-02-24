import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { Code, ConnectError } from '@connectrpc/connect'
import { useMutation, useQuery } from '@connectrpc/connect-query'
import {
  cancelSync as cancelSyncMethod,
  getSyncProgress as getSyncProgressMethod,
  inspectRoots as inspectRootsMethod,
  startSync as startSyncMethod,
} from '@/gen/npan/v1/api-AdminService_connectquery'
import {
  fromProtoGetSyncProgressResponse,
  fromProtoInspectRootsResponse,
  toProtoSyncMode,
} from '@/lib/connect-admin-adapter'
import { createNpanTransport } from '@/lib/connect-transport'
import {
  SyncProgressSchema,
  InspectRootsResponseSchema,
  preferTimestampMillis,
} from '@/lib/sync-schemas'
import type { SyncProgress, InspectRootsResponse } from '@/lib/sync-schemas'

const POLL_INTERVAL = 2000

function toErrorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    return err.rawMessage || err.message
  }
  if (err instanceof Error) {
    return err.message
  }
  return 'Unknown error'
}

function normalizeCrawlStatsTimestamps(
  stats: SyncProgress['aggregateStats'],
): SyncProgress['aggregateStats'] {
  return {
    ...stats,
    startedAt: preferTimestampMillis(stats.startedAt, stats.startedAtTs),
    endedAt: preferTimestampMillis(stats.endedAt, stats.endedAtTs),
  }
}

function normalizeRootProgressTimestamps(
  rootProgress: SyncProgress['rootProgress'],
): SyncProgress['rootProgress'] {
  const next: SyncProgress['rootProgress'] = {}
  for (const [key, value] of Object.entries(rootProgress)) {
    next[key] = {
      ...value,
      updatedAt: preferTimestampMillis(value.updatedAt, value.updatedAtTs),
      stats: normalizeCrawlStatsTimestamps(value.stats),
    }
  }
  return next
}

export function normalizeSyncProgressTimestamps(progress: SyncProgress): SyncProgress {
  return {
    ...progress,
    startedAt: preferTimestampMillis(progress.startedAt, progress.startedAtTs),
    updatedAt: preferTimestampMillis(progress.updatedAt, progress.updatedAtTs),
    aggregateStats: normalizeCrawlStatsTimestamps(progress.aggregateStats),
    rootProgress: normalizeRootProgressTimestamps(progress.rootProgress),
    catalogRootProgress: progress.catalogRootProgress
      ? normalizeRootProgressTimestamps(progress.catalogRootProgress)
      : progress.catalogRootProgress,
  }
}

export function useSyncProgress(headers: Record<string, string>) {
  const [progress, setProgress] = useState<SyncProgress | null>(null)
  const [loading, setLoading] = useState(false)
  const [initialLoading, setInitialLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [inspectLoading, setInspectLoading] = useState(false)
  const [inspectError, setInspectError] = useState<string | null>(null)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const apiKey = headers['X-API-Key']
  const transport = useMemo(
    () =>
      apiKey
        ? createNpanTransport({
            'X-API-Key': apiKey,
          })
        : undefined,
    [apiKey],
  )

  const progressQuery = useQuery(getSyncProgressMethod, {}, {
    enabled: false,
    transport,
    retry: false,
  })
  const startSyncMutation = useMutation(startSyncMethod, {
    transport,
    retry: false,
  })
  const inspectRootsMutation = useMutation(inspectRootsMethod, {
    transport,
    retry: false,
  })
  const cancelSyncMutation = useMutation(cancelSyncMethod, {
    transport,
    retry: false,
  })

  const fetchProgress = useCallback(async () => {
    try {
      const result = await progressQuery.refetch()
      if (result.error) {
        throw result.error
      }
      const mapped = result.data
        ? fromProtoGetSyncProgressResponse(result.data)
        : null
      if (!mapped) {
        setProgress(null)
        setError(null)
        return null
      }
      const normalized = normalizeSyncProgressTimestamps(
        SyncProgressSchema.parse(mapped),
      )
      setProgress(normalized)
      setError(null)
      return normalized
    } catch (err) {
      if (err instanceof ConnectError && err.code === Code.NotFound) {
        setProgress(null)
        setError(null)
        return null
      }
      setError(toErrorMessage(err))
      return null
    }
  }, [progressQuery.refetch])

  const startPolling = useCallback(
    (status: string, gracePollCount: number = 0) => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }

      if (status === 'running') {
        let remainingGrace = gracePollCount
        intervalRef.current = setInterval(() => {
          void fetchProgress().then((result) => {
            if (result && result.status !== 'running') {
              if (remainingGrace > 0) {
                remainingGrace--
                return
              }
              if (intervalRef.current) {
                clearInterval(intervalRef.current)
                intervalRef.current = null
              }
            }
          })
        }, POLL_INTERVAL)
      }
    },
    [fetchProgress],
  )

  const startSync = useCallback(
    async (
      rootFolderIds: number[],
      mode: string = 'auto',
      forceRebuild: boolean = false,
      options?: { preserveRootCatalog?: boolean },
    ) => {
      setLoading(true)
      setError(null)
      try {
        const preserveRootCatalog =
          options?.preserveRootCatalog ?? rootFolderIds.length > 0
        await startSyncMutation.mutateAsync({
          mode: toProtoSyncMode(mode),
          rootFolderIds: rootFolderIds.map((id) => BigInt(id)),
          includeDepartments: rootFolderIds.length > 0 ? false : undefined,
          preserveRootCatalog:
            preserveRootCatalog && !forceRebuild ? true : undefined,
          resumeProgress: mode !== 'full' && !forceRebuild,
          forceRebuild: forceRebuild || undefined,
        })
        await fetchProgress()
        setProgress((prev) =>
          prev
            ? {
                ...prev,
                status: 'running',
                lastError: undefined,
                updatedAt: Date.now(),
              }
            : ({
                status: 'running',
                mode: mode as 'auto' | 'full' | 'incremental',
                startedAt: Date.now(),
                updatedAt: Date.now(),
                roots: [],
                completedRoots: [],
                rootProgress: {},
                aggregateStats: {
                  filesIndexed: 0,
                  filesDiscovered: 0,
                  skippedFiles: 0,
                  pagesFetched: 0,
                  foldersVisited: 0,
                  failedRequests: 0,
                  startedAt: 0,
                  endedAt: 0,
                },
              } satisfies SyncProgress),
        )
        startPolling('running', 5)
      } catch (err) {
        setError(toErrorMessage(err))
      } finally {
        setLoading(false)
      }
    },
    [startSyncMutation, fetchProgress, startPolling],
  )

  const inspectRoots = useCallback(
    async (folderIds: number[]): Promise<InspectRootsResponse | null> => {
      setInspectLoading(true)
      setInspectError(null)
      try {
        const result = await inspectRootsMutation.mutateAsync({
          folderIds: folderIds.map((id) => BigInt(id)),
        })
        const parsed = InspectRootsResponseSchema.parse(
          fromProtoInspectRootsResponse(result),
        )
        if (parsed.items.length > 0) {
          setProgress((prev) => {
            const now = Date.now()
            const base: SyncProgress =
              prev ??
              ({
                status: 'idle',
                mode: 'full',
                startedAt: now,
                updatedAt: now,
                roots: [],
                completedRoots: [],
                rootProgress: {},
                aggregateStats: {
                  filesIndexed: 0,
                  filesDiscovered: 0,
                  skippedFiles: 0,
                  pagesFetched: 0,
                  foldersVisited: 0,
                  failedRequests: 0,
                  startedAt: 0,
                  endedAt: 0,
                },
              } satisfies SyncProgress)

            const catalogRootProgress = {
              ...(base.catalogRootProgress ?? base.rootProgress),
            }
            const catalogRootNames = {
              ...(base.catalogRootNames ?? base.rootNames),
            }
            const rootIDs = new Set<number>(base.catalogRoots ?? [])

            for (const item of parsed.items) {
              const key = String(item.folder_id)
              rootIDs.add(item.folder_id)
              catalogRootNames[String(item.folder_id)] = item.name
              const existing = catalogRootProgress[key]
              if (existing) {
                catalogRootProgress[key] = {
                  ...existing,
                  estimatedTotalDocs: item.estimated_total_docs,
                  updatedAt: now,
                }
                continue
              }
              catalogRootProgress[key] = {
                rootFolderId: item.folder_id,
                status: 'pending',
                estimatedTotalDocs: item.estimated_total_docs,
                stats: {
                  foldersVisited: 0,
                  filesIndexed: 0,
                  filesDiscovered: 0,
                  skippedFiles: 0,
                  pagesFetched: 0,
                  failedRequests: 0,
                  startedAt: 0,
                  endedAt: 0,
                },
                updatedAt: now,
              }
            }

            return {
              ...base,
              updatedAt: now,
              catalogRoots: [...rootIDs].sort((a, b) => a - b),
              catalogRootNames,
              catalogRootProgress,
            }
          })
        }
        return parsed
      } catch (err) {
        setInspectError(toErrorMessage(err))
        return null
      } finally {
        setInspectLoading(false)
      }
    },
    [inspectRootsMutation],
  )

  const cancelSync = useCallback(async () => {
    setLoading(true)
    try {
      await cancelSyncMutation.mutateAsync({})
      await fetchProgress()
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    } catch (err) {
      setError(toErrorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [cancelSyncMutation, fetchProgress])

  const hasAuth = Boolean(apiKey)
  useEffect(() => {
    if (!hasAuth) {
      setInitialLoading(false)
      return
    }

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

  return {
    progress,
    loading,
    initialLoading,
    error,
    inspectLoading,
    inspectError,
    startSync,
    inspectRoots,
    cancelSync,
    refetch: fetchProgress,
  }
}
