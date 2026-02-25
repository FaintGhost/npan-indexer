import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { createClient, Code, ConnectError } from '@connectrpc/connect'
import { useMutation, useQuery } from '@connectrpc/connect-query'
import {
  cancelSync as cancelSyncMethod,
  getIndexStats as getIndexStatsMethod,
  getSyncProgress as getSyncProgressMethod,
  inspectRoots as inspectRootsMethod,
  startSync as startSyncMethod,
} from '@/gen/npan/v1/api-AdminService_connectquery'
import { AdminService } from '@/gen/npan/v1/api_pb'
import {
  fromProtoGetIndexStatsResponse,
  fromProtoGetSyncProgressResponse,
  fromProtoInspectRootsResponse,
  fromProtoSyncProgressState,
  toProtoSyncMode,
} from '@/lib/connect-admin-adapter'
import { createNpanTransport } from '@/lib/connect-transport'
import {
  InspectRootsResponseSchema,
  SyncProgressSchema,
  preferTimestampMillis,
} from '@/lib/sync-schemas'
import type { InspectRootsResponse, SyncProgress } from '@/lib/sync-schemas'
import { useAdminAuth } from '@/hooks/use-admin-auth'
import { ApiKeyDialog } from '@/components/api-key-dialog'
import { SyncProgressDisplay } from '@/components/sync-progress-display'
import { ConfirmDialog } from '@/components/confirm-dialog'

const POLL_INTERVAL = 2000
const STREAM_RECONNECT_DELAY = 500

type SyncModeValue = 'full' | 'incremental'
type IndexState = 'checking' | 'ready' | 'empty' | 'unknown'

const SYNC_MODES = [
  { value: 'full', label: '全量', description: '重新爬取所有目录' },
  { value: 'incremental', label: '增量', description: '仅同步新增、更新、删除' },
] as const

function getSelectableRootIDs(progress: SyncProgress | null): number[] {
  if (!progress) return []

  const fromCatalog = progress.catalogRoots ?? []
  if (fromCatalog.length > 0) {
    return [...new Set(fromCatalog)].sort((a, b) => a - b)
  }

  const source = progress.catalogRootProgress ?? progress.rootProgress
  const ids = Object.keys(source ?? {})
    .map((key) => Number(key))
    .filter((id) => Number.isInteger(id) && id > 0)

  return [...new Set(ids)].sort((a, b) => a - b)
}

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

function normalizeSyncProgressTimestamps(progress: SyncProgress): SyncProgress {
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

function getIncrementalBlockedReason(indexState: IndexState): string {
  switch (indexState) {
    case 'checking':
      return '正在检查索引状态...'
    case 'empty':
      return '请先执行一次全量索引'
    case 'unknown':
      return '无法确认索引状态，请稍后重试'
    default:
      return ''
  }
}

function getIndexStateHint(indexState: IndexState): string {
  switch (indexState) {
    case 'ready':
      return '索引状态正常，可执行增量同步'
    case 'checking':
      return '正在检查索引状态...'
    case 'empty':
      return '请先执行一次全量索引'
    case 'unknown':
    default:
      return '无法确认索引状态，请稍后重试'
  }
}

export function AdminSyncPage() {
  const auth = useAdminAuth()
  const [progress, setProgress] = useState<SyncProgress | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [inspectError, setInspectError] = useState<string | null>(null)
  const [message, setMessage] = useState<string | null>(null)
  const [mode, setMode] = useState<SyncModeValue>('full')
  const [forceRebuild, setForceRebuild] = useState(false)
  const [selectedRootIDs, setSelectedRootIDs] = useState<number[]>([])
  const knownRootIDsRef = useRef<Set<number>>(new Set())
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean
    title: string
    message: string
    confirmLabel: string
    variant: 'danger' | 'default'
    onConfirm: () => void
  }>({
    open: false,
    title: '',
    message: '',
    confirmLabel: '',
    variant: 'default',
    onConfirm: () => {},
  })

  const apiKey = auth.apiKey
  const onUnauthorized = auth.on401
  const hasAuth = Boolean(apiKey)
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
    enabled: hasAuth,
    transport,
    retry: false,
    select: (response) => {
      const mapped = fromProtoGetSyncProgressResponse(response)
      if (!mapped) {
        return null
      }
      return normalizeSyncProgressTimestamps(SyncProgressSchema.parse(mapped))
    },
    // Keep a polling fallback even when current data is null/not_found.
    // Streaming should provide realtime updates, but this prevents the UI from
    // getting stuck if the stream is unavailable or temporarily broken.
    refetchInterval: POLL_INTERVAL,
  })
  const indexStatsQuery = useQuery(getIndexStatsMethod, {}, {
    enabled: hasAuth,
    transport,
    retry: false,
    select: (response) => fromProtoGetIndexStatsResponse(response),
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

  useEffect(() => {
    if (!hasAuth) {
      setProgress(null)
      setError(null)
      setInspectError(null)
      setSelectedRootIDs([])
      knownRootIDsRef.current = new Set()
    }
  }, [hasAuth])

  useEffect(() => {
    const err = indexStatsQuery.error
    if (err instanceof ConnectError && err.code === Code.Unauthenticated) {
      onUnauthorized()
    }
  }, [indexStatsQuery.error, onUnauthorized])

  useEffect(() => {
    if (progressQuery.data === undefined) {
      return
    }
    setProgress(progressQuery.data ?? null)
    setError(null)
  }, [progressQuery.data])

  useEffect(() => {
    const queryError = progressQuery.error
    if (!queryError) {
      return
    }

    if (queryError instanceof ConnectError && queryError.code === Code.NotFound) {
      setProgress(null)
      setError(null)
      return
    }

    setError(toErrorMessage(queryError))
  }, [progressQuery.error])

  const adminStreamClient = useMemo(
    () => (transport ? createClient(AdminService, transport) : undefined),
    [transport],
  )

  useEffect(() => {
    if (!hasAuth || !adminStreamClient) {
      return
    }

    const abortController = new AbortController()
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null

    const scheduleReconnect = (fn: () => void) => {
      reconnectTimer = setTimeout(fn, STREAM_RECONNECT_DELAY)
    }

    const watchProgress = async () => {
      try {
        for await (const response of adminStreamClient.watchSyncProgress(
          {},
          { signal: abortController.signal },
        )) {
          if (!response.state) {
            continue
          }
          const mapped = fromProtoSyncProgressState(response.state)
          const normalized = normalizeSyncProgressTimestamps(
            SyncProgressSchema.parse(mapped),
          )
          setProgress(normalized)
          setError(null)
        }

        if (!abortController.signal.aborted) {
          scheduleReconnect(() => {
            void watchProgress()
          })
        }
      } catch (err) {
        if (abortController.signal.aborted) {
          return
        }
        if (err instanceof ConnectError) {
          if (err.code === Code.Unauthenticated) {
            onUnauthorized()
            return
          }
          if (err.code === Code.Unimplemented) {
            return
          }
          if (err.code === Code.NotFound) {
            setProgress(null)
            setError(null)
          }
        }
        scheduleReconnect(() => {
          void watchProgress()
        })
      }
    }

    void watchProgress()

    return () => {
      abortController.abort()
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
      }
    }
  }, [adminStreamClient, hasAuth, onUnauthorized])

  const startSync = useCallback(
    async (
      rootFolderIds: number[],
      selectedMode: SyncModeValue = 'full',
      selectedForceRebuild: boolean = false,
      options?: { preserveRootCatalog?: boolean },
    ): Promise<boolean> => {
      setError(null)
      try {
        const preserveRootCatalog =
          options?.preserveRootCatalog ?? rootFolderIds.length > 0
        await startSyncMutation.mutateAsync({
          mode: toProtoSyncMode(selectedMode),
          rootFolderIds: rootFolderIds.map((id) => BigInt(id)),
          includeDepartments: rootFolderIds.length > 0 ? false : undefined,
          preserveRootCatalog:
            preserveRootCatalog && !selectedForceRebuild ? true : undefined,
          // 默认允许断点续传；仅在强制重建时显式关闭。
          resumeProgress: selectedForceRebuild ? false : undefined,
          forceRebuild: selectedForceRebuild || undefined,
        })
        await progressQuery.refetch()
        setProgress((prev) =>
          prev
            ? {
                ...prev,
                status: 'running',
                lastError: undefined,
                updatedAt: Date.now(),
              }
            : {
                status: 'running',
                mode: selectedMode,
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
              },
        )
        return true
      } catch (err) {
        setError(toErrorMessage(err))
        return false
      }
    },
    [progressQuery, startSyncMutation],
  )

  const inspectRoots = useCallback(
    async (folderIds: number[]): Promise<InspectRootsResponse | null> => {
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
              progressQuery.data ?? {
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
              }

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
      }
    },
    [inspectRootsMutation, progressQuery.data],
  )

  const cancelSync = useCallback(async (): Promise<boolean> => {
    setError(null)
    try {
      await cancelSyncMutation.mutateAsync({})
      await progressQuery.refetch()
      return true
    } catch (err) {
      setError(toErrorMessage(err))
      return false
    }
  }, [cancelSyncMutation, progressQuery])

  const loading = startSyncMutation.isPending || cancelSyncMutation.isPending
  const inspectLoading = inspectRootsMutation.isPending
  const initialLoading = hasAuth && progressQuery.isPending && progress === null
  const isRunning = progress?.status === 'running'
  const indexState: IndexState = useMemo(() => {
    if (!hasAuth || indexStatsQuery.isPending) {
      return 'checking'
    }
    if (indexStatsQuery.error) {
      return 'unknown'
    }
    const docCount = indexStatsQuery.data
    if (docCount == null) {
      return 'checking'
    }
    return docCount === 0 ? 'empty' : 'ready'
  }, [hasAuth, indexStatsQuery.data, indexStatsQuery.error, indexStatsQuery.isPending])
  const canSwitchMode = !isRunning && !loading
  const canStartSync = !isRunning && !loading && (mode !== 'incremental' || indexState === 'ready')
  const canInspectRoots = !inspectLoading
  const selectableRootIDs = useMemo(
    () => getSelectableRootIDs(progress),
    [progress],
  )
  const selectedScopedRoots = useMemo(() => {
    if (mode !== 'full') return []
    const selected = new Set(selectedRootIDs)
    return selectableRootIDs.filter((id) => selected.has(id))
  }, [mode, selectableRootIDs, selectedRootIDs])

  useEffect(() => {
    if (selectableRootIDs.length === 0) return
    setSelectedRootIDs((prev) => {
      const next = new Set(prev)
      let changed = false
      for (const rootID of selectableRootIDs) {
        if (knownRootIDsRef.current.has(rootID)) continue
        knownRootIDsRef.current.add(rootID)
        next.add(rootID)
        changed = true
      }
      if (!changed) return prev
      return [...next].sort((a, b) => a - b)
    })
  }, [selectableRootIDs])

  const handleInspectRoots = async () => {
    setMessage(null)
    if (selectableRootIDs.length === 0) {
      setMessage('暂无可拉取的根目录，请先完成一次全量同步')
      return
    }

    const result = await inspectRoots(selectableRootIDs)
    if (!result) return

    const successCount = result.items.length
    const failCount = result.errors?.length ?? 0
    setMessage(
      failCount > 0
        ? `目录详情已拉取：成功 ${successCount} 个，失败 ${failCount} 个`
        : `目录详情已拉取：成功 ${successCount} 个`,
    )
    setTimeout(() => setMessage(null), 4000)
  }

  const handleStartSync = async () => {
    if (isRunning) {
      setMessage('当前已有同步任务运行中，请先取消后再启动')
      return
    }

    if (mode === 'incremental' && indexState !== 'ready') {
      setMessage(getIncrementalBlockedReason(indexState))
      return
    }

    if (forceRebuild && selectedScopedRoots.length > 0) {
      setMessage('强制重建仅允许全量全库执行，请先取消勾选目录')
      return
    }

    if (forceRebuild) {
      setConfirmDialog({
        open: true,
        title: '强制重建索引',
        message:
          '此操作将清空所有索引数据并重新爬取，重建期间搜索将无结果。确认继续？',
        confirmLabel: '确认重建',
        variant: 'danger',
        onConfirm: () => {
          setConfirmDialog((prev) => ({ ...prev, open: false }))
          void doStartSync()
        },
      })
      return
    }
    await doStartSync()
  }

  const doStartSync = async () => {
    setMessage(null)
    const scopedRootIDs = mode === 'full' ? selectedScopedRoots : []
    const started = await startSync(scopedRootIDs, mode, forceRebuild, {
      preserveRootCatalog: scopedRootIDs.length > 0,
    })
    if (started) {
      setMessage('同步任务已启动')
      setTimeout(() => setMessage(null), 4000)
    }
  }

  const handleCancelSync = async () => {
    setConfirmDialog({
      open: true,
      title: '取消同步',
      message: '确认取消当前正在进行的同步任务？',
      confirmLabel: '确认取消',
      variant: 'danger',
      onConfirm: () => {
        setConfirmDialog((prev) => ({ ...prev, open: false }))
        void doCancelSync()
      },
    })
  }

  const doCancelSync = async () => {
    setMessage(null)
    const cancelled = await cancelSync()
    if (cancelled) {
      setMessage('已发送取消请求')
      setTimeout(() => setMessage(null), 4000)
    }
  }

  if (auth.needsAuth) {
    return (
      <ApiKeyDialog
        open
        onSubmit={(key) => auth.validate(key)}
        error={auth.error}
        loading={auth.loading}
      />
    )
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <div className="mb-8 flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-slate-900">同步管理</h1>
        <a
          href="/"
          className="text-sm text-slate-500 transition-colors hover:text-slate-700"
        >
          ← 返回搜索
        </a>
      </div>

      <div className="mb-6 space-y-3">
        <div className="flex gap-1 rounded-lg bg-slate-100 p-1">
          {SYNC_MODES.map((m) => (
            <button
              key={m.value}
              type="button"
              onClick={() => setMode(m.value)}
              disabled={
                !canSwitchMode || (m.value === 'incremental' && indexState !== 'ready')
              }
              className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                mode === m.value
                  ? 'bg-white text-slate-900 shadow-sm'
                  : 'text-slate-500 hover:text-slate-700'
              } disabled:cursor-not-allowed disabled:opacity-60`}
              title={m.description}
            >
              {m.label}
            </button>
          ))}
        </div>

        <div className="space-y-2 rounded-xl border border-slate-200 bg-white p-4">
          <p className="block text-sm font-medium text-slate-700">
            根目录详情
          </p>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleInspectRoots}
              disabled={!canInspectRoots}
              className="shrink-0 rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {inspectLoading ? '拉取中...' : '刷新目录详情'}
            </button>
          </div>
          <p className="text-xs text-slate-400">
            该操作仅刷新已存在根目录的详情，不会启动同步。
          </p>
          {selectableRootIDs.length === 0 && (
            <p className="text-xs text-amber-600">
              当前没有可刷新目录，请先完成一次全量同步以生成根目录列表。
            </p>
          )}
          {mode === 'full' && selectableRootIDs.length > 0 && (
            <p className="text-xs text-slate-500">
              当前已勾选 {selectedScopedRoots.length} / {selectableRootIDs.length} 个根目录；启动全量时将仅同步勾选目录。
            </p>
          )}
        </div>

        {!isRunning && mode === 'full' && (
          <button
            type="button"
            role="switch"
            aria-checked={forceRebuild}
            onClick={() => setForceRebuild(!forceRebuild)}
            disabled={isRunning || loading}
            className="flex items-center gap-3 rounded-xl border border-slate-200 bg-white px-4 py-3 text-left transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <span
              className={`relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors ${
                forceRebuild ? 'bg-rose-500' : 'bg-slate-200'
              }`}
            >
              <span
                className={`inline-block h-3.5 w-3.5 rounded-full bg-white shadow-sm transition-transform ${
                  forceRebuild ? 'translate-x-4' : 'translate-x-1'
                }`}
              />
            </span>
            <span className="flex flex-col">
              <span className="text-sm font-medium text-slate-700">
                强制重建索引
              </span>
              <span className="text-xs text-slate-400">
                仅允许全量全库执行；会清空现有索引后重置断点重新爬取
              </span>
            </span>
          </button>
        )}

        <div className="flex gap-3">
          <button
            type="button"
            onClick={handleStartSync}
            disabled={!canStartSync}
            className="inline-flex items-center gap-2 rounded-xl bg-slate-900 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {loading && (
              <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
            )}
            启动同步
          </button>

          {isRunning && (
            <button
              type="button"
              onClick={handleCancelSync}
              disabled={loading}
              className="rounded-xl border border-rose-200 bg-white px-5 py-2.5 text-sm font-medium text-rose-600 transition-colors hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-60"
            >
              取消同步
            </button>
          )}
        </div>

        <div className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">
          <p
            className={`text-xs ${
              indexState === 'ready' ? 'text-slate-600' : 'text-amber-700'
            }`}
          >
            {getIndexStateHint(indexState)}
          </p>
        </div>
      </div>

      {message && (
        <div className="mb-4 rounded-xl border border-emerald-200 bg-emerald-50 p-3">
          <p className="text-sm text-emerald-700">{message}</p>
        </div>
      )}

      {inspectError && (
        <div className="mb-4 rounded-xl border border-amber-200 bg-amber-50 p-3">
          <p className="text-sm text-amber-700">{inspectError}</p>
        </div>
      )}

      {error && (
        <div className="mb-4 rounded-xl border border-rose-200 bg-rose-50 p-3">
          <p className="text-sm text-rose-600">{error}</p>
        </div>
      )}

      {initialLoading && (
        <div className="space-y-4">
          <div className="h-8 w-24 animate-pulse rounded-lg bg-slate-200" />
          <div className="h-24 animate-pulse rounded-xl bg-slate-100" />
          <div className="grid grid-cols-2 gap-3">
            <div className="h-20 animate-pulse rounded-xl bg-slate-100" />
            <div className="h-20 animate-pulse rounded-xl bg-slate-100" />
          </div>
        </div>
      )}

      {!initialLoading && progress && (
        <SyncProgressDisplay
          progress={progress}
          rootSelection={
            mode === 'full'
              ? {
                  selectedRootIds: selectedRootIDs,
                  disabled: isRunning || loading,
                  onToggleRoot: (rootID) => {
                    setSelectedRootIDs((prev) =>
                      prev.includes(rootID)
                        ? prev.filter((id) => id !== rootID)
                        : [...prev, rootID].sort((a, b) => a - b),
                    )
                  },
                }
              : undefined
          }
        />
      )}

      {!initialLoading && !progress && !loading && !error && (
        <div className="py-12 text-center">
          <p className="text-sm text-slate-400">暂无同步记录</p>
        </div>
      )}

      <ConfirmDialog
        open={confirmDialog.open}
        title={confirmDialog.title}
        message={confirmDialog.message}
        confirmLabel={confirmDialog.confirmLabel}
        variant={confirmDialog.variant}
        onConfirm={confirmDialog.onConfirm}
        onCancel={() => setConfirmDialog((prev) => ({ ...prev, open: false }))}
      />
    </div>
  )
}
