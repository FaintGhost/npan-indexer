import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Code, ConnectError } from "@connectrpc/connect";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  cancelSync as cancelSyncMethod,
  getSyncProgress as getSyncProgressMethod,
  inspectRoots as inspectRootsMethod,
  startSync as startSyncMethod,
} from "@/gen/npan/v1/api-AdminService_connectquery";
import {
  fromProtoGetSyncProgressResponse,
  fromProtoInspectRootsResponse,
  toProtoSyncMode,
} from "@/lib/connect-admin-adapter";
import { createNpanTransport } from "@/lib/connect-transport";
import {
  InspectRootsResponseSchema,
  SyncProgressSchema,
  preferTimestampMillis,
} from "@/lib/sync-schemas";
import type { InspectRootsResponse, SyncProgress } from "@/lib/sync-schemas";
import { useAdminAuth } from "@/hooks/use-admin-auth";
import { ApiKeyDialog } from "@/components/api-key-dialog";
import { SyncProgressDisplay } from "@/components/sync-progress-display";
import { ConfirmDialog } from "@/components/confirm-dialog";

const POLL_INTERVAL = 2000;

const SYNC_MODES = [
  { value: "auto", label: "自适应", description: "有游标走增量，否则全量" },
  { value: "full", label: "全量", description: "重新爬取所有目录" },
  { value: "incremental", label: "增量", description: "仅同步最近变更" },
] as const;

function getSelectableRootIDs(progress: SyncProgress | null): number[] {
  if (!progress) return [];

  const fromCatalog = progress.catalogRoots ?? [];
  if (fromCatalog.length > 0) {
    return [...new Set(fromCatalog)].sort((a, b) => a - b);
  }

  const source = progress.catalogRootProgress ?? progress.rootProgress;
  const ids = Object.keys(source ?? {})
    .map((key) => Number(key))
    .filter((id) => Number.isInteger(id) && id > 0);

  return [...new Set(ids)].sort((a, b) => a - b);
}

function toErrorMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    return err.rawMessage || err.message;
  }
  if (err instanceof Error) {
    return err.message;
  }
  return "Unknown error";
}

function normalizeCrawlStatsTimestamps(
  stats: SyncProgress["aggregateStats"],
): SyncProgress["aggregateStats"] {
  return {
    ...stats,
    startedAt: preferTimestampMillis(stats.startedAt, stats.startedAtTs),
    endedAt: preferTimestampMillis(stats.endedAt, stats.endedAtTs),
  };
}

function normalizeRootProgressTimestamps(
  rootProgress: SyncProgress["rootProgress"],
): SyncProgress["rootProgress"] {
  const next: SyncProgress["rootProgress"] = {};
  for (const [key, value] of Object.entries(rootProgress)) {
    next[key] = {
      ...value,
      updatedAt: preferTimestampMillis(value.updatedAt, value.updatedAtTs),
      stats: normalizeCrawlStatsTimestamps(value.stats),
    };
  }
  return next;
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
  };
}

export function AdminSyncPage() {
  const auth = useAdminAuth();
  const [progress, setProgress] = useState<SyncProgress | null>(null);
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [inspectLoading, setInspectLoading] = useState(false);
  const [inspectError, setInspectError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [mode, setMode] = useState<string>("auto");
  const [forceRebuild, setForceRebuild] = useState(false);
  const [selectedRootIDs, setSelectedRootIDs] = useState<number[]>([]);
  const selectionInitializedRef = useRef(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean;
    title: string;
    message: string;
    confirmLabel: string;
    variant: "danger" | "default";
    onConfirm: () => void;
  }>({
    open: false,
    title: "",
    message: "",
    confirmLabel: "",
    variant: "default",
    onConfirm: () => {},
  });

  const apiKey = auth.apiKey;
  const hasAuth = Boolean(apiKey);
  const transport = useMemo(
    () =>
      apiKey
        ? createNpanTransport({
            "X-API-Key": apiKey,
          })
        : undefined,
    [apiKey],
  );

  const progressQuery = useQuery(getSyncProgressMethod, {}, {
    enabled: false,
    transport,
    retry: false,
  });
  const startSyncMutation = useMutation(startSyncMethod, {
    transport,
    retry: false,
  });
  const inspectRootsMutation = useMutation(inspectRootsMethod, {
    transport,
    retry: false,
  });
  const cancelSyncMutation = useMutation(cancelSyncMethod, {
    transport,
    retry: false,
  });

  const fetchProgress = useCallback(async (): Promise<SyncProgress | null> => {
    try {
      const result = await progressQuery.refetch();
      if (result.error) {
        throw result.error;
      }
      const mapped = result.data
        ? fromProtoGetSyncProgressResponse(result.data)
        : null;
      if (!mapped) {
        setProgress(null);
        setError(null);
        return null;
      }
      const normalized = normalizeSyncProgressTimestamps(
        SyncProgressSchema.parse(mapped),
      );
      setProgress(normalized);
      setError(null);
      return normalized;
    } catch (err) {
      if (err instanceof ConnectError && err.code === Code.NotFound) {
        setProgress(null);
        setError(null);
        return null;
      }
      setError(toErrorMessage(err));
      return null;
    }
  }, [progressQuery.refetch]);

  const startPolling = useCallback(
    (status: string, gracePollCount: number = 0) => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }

      if (status === "running") {
        let remainingGrace = gracePollCount;
        intervalRef.current = setInterval(() => {
          void fetchProgress().then((result) => {
            if (result && result.status !== "running") {
              if (remainingGrace > 0) {
                remainingGrace--;
                return;
              }
              if (intervalRef.current) {
                clearInterval(intervalRef.current);
                intervalRef.current = null;
              }
            }
          });
        }, POLL_INTERVAL);
      }
    },
    [fetchProgress],
  );

  const startSync = useCallback(
    async (
      rootFolderIds: number[],
      selectedMode: string = "auto",
      selectedForceRebuild: boolean = false,
      options?: { preserveRootCatalog?: boolean },
    ): Promise<boolean> => {
      setLoading(true);
      setError(null);
      try {
        const preserveRootCatalog =
          options?.preserveRootCatalog ?? rootFolderIds.length > 0;
        await startSyncMutation.mutateAsync({
          mode: toProtoSyncMode(selectedMode),
          rootFolderIds: rootFolderIds.map((id) => BigInt(id)),
          includeDepartments: rootFolderIds.length > 0 ? false : undefined,
          preserveRootCatalog:
            preserveRootCatalog && !selectedForceRebuild ? true : undefined,
          resumeProgress: selectedMode !== "full" && !selectedForceRebuild,
          forceRebuild: selectedForceRebuild || undefined,
        });
        await fetchProgress();
        setProgress((prev) =>
          prev
            ? {
                ...prev,
                status: "running",
                lastError: undefined,
                updatedAt: Date.now(),
              }
            : {
                status: "running",
                mode: selectedMode as "auto" | "full" | "incremental",
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
        );
        startPolling("running", 5);
        return true;
      } catch (err) {
        setError(toErrorMessage(err));
        return false;
      } finally {
        setLoading(false);
      }
    },
    [startSyncMutation, fetchProgress, startPolling],
  );

  const inspectRoots = useCallback(
    async (folderIds: number[]): Promise<InspectRootsResponse | null> => {
      setInspectLoading(true);
      setInspectError(null);
      try {
        const result = await inspectRootsMutation.mutateAsync({
          folderIds: folderIds.map((id) => BigInt(id)),
        });
        const parsed = InspectRootsResponseSchema.parse(
          fromProtoInspectRootsResponse(result),
        );
        if (parsed.items.length > 0) {
          setProgress((prev) => {
            const now = Date.now();
            const base: SyncProgress =
              prev ?? {
                status: "idle",
                mode: "full",
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
              };

            const catalogRootProgress = {
              ...(base.catalogRootProgress ?? base.rootProgress),
            };
            const catalogRootNames = {
              ...(base.catalogRootNames ?? base.rootNames),
            };
            const rootIDs = new Set<number>(base.catalogRoots ?? []);

            for (const item of parsed.items) {
              const key = String(item.folder_id);
              rootIDs.add(item.folder_id);
              catalogRootNames[String(item.folder_id)] = item.name;
              const existing = catalogRootProgress[key];
              if (existing) {
                catalogRootProgress[key] = {
                  ...existing,
                  estimatedTotalDocs: item.estimated_total_docs,
                  updatedAt: now,
                };
                continue;
              }
              catalogRootProgress[key] = {
                rootFolderId: item.folder_id,
                status: "pending",
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
              };
            }

            return {
              ...base,
              updatedAt: now,
              catalogRoots: [...rootIDs].sort((a, b) => a - b),
              catalogRootNames,
              catalogRootProgress,
            };
          });
        }
        return parsed;
      } catch (err) {
        setInspectError(toErrorMessage(err));
        return null;
      } finally {
        setInspectLoading(false);
      }
    },
    [inspectRootsMutation],
  );

  const cancelSync = useCallback(async (): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      await cancelSyncMutation.mutateAsync({});
      await fetchProgress();
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return true;
    } catch (err) {
      setError(toErrorMessage(err));
      return false;
    } finally {
      setLoading(false);
    }
  }, [cancelSyncMutation, fetchProgress]);

  useEffect(() => {
    if (!hasAuth) {
      setInitialLoading(false);
      setProgress(null);
      setError(null);
      return;
    }

    setInitialLoading(true);
    void fetchProgress().then((result) => {
      setInitialLoading(false);
      if (result) {
        startPolling(result.status);
      }
    });

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [hasAuth]); // eslint-disable-line react-hooks/exhaustive-deps

  const isRunning = progress?.status === "running";
  const isBusy = loading || inspectLoading || isRunning;
  const selectableRootIDs = useMemo(
    () => getSelectableRootIDs(progress),
    [progress],
  );
  const selectedScopedRoots = useMemo(() => {
    if (mode !== "full") return [];
    const selected = new Set(selectedRootIDs);
    return selectableRootIDs.filter((id) => selected.has(id));
  }, [mode, selectableRootIDs, selectedRootIDs]);

  useEffect(() => {
    if (selectionInitializedRef.current) return;
    if (selectableRootIDs.length === 0) return;
    setSelectedRootIDs(selectableRootIDs);
    selectionInitializedRef.current = true;
  }, [selectableRootIDs]);

  const handleInspectRoots = async () => {
    setMessage(null);
    if (selectableRootIDs.length === 0) {
      setMessage("暂无可拉取的根目录，请先完成一次全量同步");
      return;
    }

    const result = await inspectRoots(selectableRootIDs);
    if (!result) return;

    const successCount = result.items.length;
    const failCount = result.errors?.length ?? 0;
    setMessage(
      failCount > 0
        ? `目录详情已拉取：成功 ${successCount} 个，失败 ${failCount} 个`
        : `目录详情已拉取：成功 ${successCount} 个`,
    );
    setTimeout(() => setMessage(null), 4000);
  };

  const handleStartSync = async () => {
    if (forceRebuild && selectedScopedRoots.length > 0) {
      setMessage("强制重建仅允许全量全库执行，请先取消勾选目录");
      return;
    }

    if (forceRebuild) {
      setConfirmDialog({
        open: true,
        title: "强制重建索引",
        message:
          "此操作将清空所有索引数据并重新爬取，重建期间搜索将无结果。确认继续？",
        confirmLabel: "确认重建",
        variant: "danger",
        onConfirm: () => {
          setConfirmDialog((prev) => ({ ...prev, open: false }));
          void doStartSync();
        },
      });
      return;
    }
    await doStartSync();
  };

  const doStartSync = async () => {
    setMessage(null);
    const scopedRootIDs = mode === "full" ? selectedScopedRoots : [];
    const started = await startSync(scopedRootIDs, mode, forceRebuild, {
      preserveRootCatalog: scopedRootIDs.length > 0,
    });
    if (started) {
      setMessage("同步任务已启动");
      setTimeout(() => setMessage(null), 4000);
    }
  };

  const handleCancelSync = async () => {
    setConfirmDialog({
      open: true,
      title: "取消同步",
      message: "确认取消当前正在进行的同步任务？",
      confirmLabel: "确认取消",
      variant: "danger",
      onConfirm: () => {
        setConfirmDialog((prev) => ({ ...prev, open: false }));
        void doCancelSync();
      },
    });
  };

  const doCancelSync = async () => {
    setMessage(null);
    const cancelled = await cancelSync();
    if (cancelled) {
      setMessage("已发送取消请求");
      setTimeout(() => setMessage(null), 4000);
    }
  };

  if (auth.needsAuth) {
    return (
      <ApiKeyDialog
        open
        onSubmit={(key) => auth.validate(key)}
        error={auth.error}
        loading={auth.loading}
      />
    );
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
        {!isRunning && (
          <div className="flex gap-1 rounded-lg bg-slate-100 p-1">
            {SYNC_MODES.map((m) => (
              <button
                key={m.value}
                type="button"
                onClick={() => setMode(m.value)}
                disabled={isBusy}
                className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                  mode === m.value
                    ? "bg-white text-slate-900 shadow-sm"
                    : "text-slate-500 hover:text-slate-700"
                } disabled:cursor-not-allowed disabled:opacity-60`}
                title={m.description}
              >
                {m.label}
              </button>
            ))}
          </div>
        )}

        {!isRunning && (
          <div className="space-y-2 rounded-xl border border-slate-200 bg-white p-4">
            <p className="block text-sm font-medium text-slate-700">
              根目录详情
            </p>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={handleInspectRoots}
                disabled={isBusy || selectableRootIDs.length === 0}
                className="shrink-0 rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {inspectLoading ? "拉取中..." : "刷新目录详情"}
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
            {mode === "full" && selectableRootIDs.length > 0 && (
              <p className="text-xs text-slate-500">
                当前已勾选 {selectedScopedRoots.length} / {selectableRootIDs.length} 个根目录；启动全量时将仅同步勾选目录。
              </p>
            )}
          </div>
        )}

        {!isRunning && mode === "full" && (
          <button
            type="button"
            role="switch"
            aria-checked={forceRebuild}
            onClick={() => setForceRebuild(!forceRebuild)}
            disabled={isBusy}
            className="flex items-center gap-3 rounded-xl border border-slate-200 bg-white px-4 py-3 text-left transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <span
              className={`relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors ${
                forceRebuild ? "bg-rose-500" : "bg-slate-200"
              }`}
            >
              <span
                className={`inline-block h-3.5 w-3.5 rounded-full bg-white shadow-sm transition-transform ${
                  forceRebuild ? "translate-x-4" : "translate-x-1"
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
            disabled={isBusy}
            className="inline-flex items-center gap-2 rounded-xl bg-slate-900 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {loading && !isRunning && (
              <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
            )}
            {isRunning && (
              <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-emerald-400" />
            )}
            {loading && !isRunning
              ? "启动中..."
              : isRunning
                ? "同步进行中"
                : mode === "full" && selectedScopedRoots.length > 0
                  ? "按勾选目录启动全量"
                  : "启动同步"}
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
          rootSelection={{
            selectedRootIds: selectedRootIDs,
            disabled: isBusy,
            onToggleRoot: (rootID) => {
              setSelectedRootIDs((prev) =>
                prev.includes(rootID)
                  ? prev.filter((id) => id !== rootID)
                  : [...prev, rootID].sort((a, b) => a - b),
              );
            },
          }}
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
  );
}
