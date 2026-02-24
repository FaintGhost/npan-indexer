import { useState, useEffect, useCallback, useRef } from "react";
import { apiGet, apiPost, apiDelete, ApiError } from "@/lib/api-client";
import {
  SyncProgressSchema,
  InspectRootsResponseSchema,
} from "@/lib/sync-schemas";
import type { SyncProgress, InspectRootsResponse } from "@/lib/sync-schemas";

const POLL_INTERVAL = 2000;

export function useSyncProgress(headers: Record<string, string>) {
  const [progress, setProgress] = useState<SyncProgress | null>(null);
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [inspectLoading, setInspectLoading] = useState(false);
  const [inspectError, setInspectError] = useState<string | null>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchProgress = useCallback(async () => {
    try {
      const result = await apiGet(
        "/api/v1/admin/sync",
        {},
        SyncProgressSchema,
        { headers },
      );
      setProgress(result as SyncProgress);
      setError(null);
      return result;
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        // No sync progress yet — not an error
        setProgress(null);
        setError(null);
        return null;
      }
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError(err instanceof Error ? err.message : "Unknown error");
      }
      return null;
    }
  }, [headers]);

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
      mode: string = "auto",
      forceRebuild: boolean = false,
      options?: { preserveRootCatalog?: boolean },
    ) => {
      setLoading(true);
      setError(null);
      try {
        const preserveRootCatalog =
          options?.preserveRootCatalog ?? (rootFolderIds.length > 0);
        await apiPost(
          "/api/v1/admin/sync",
          {
            root_folder_ids: rootFolderIds,
            include_departments:
              rootFolderIds.length > 0 ? false : undefined,
            preserve_root_catalog:
              preserveRootCatalog && !forceRebuild ? true : undefined,
            resume_progress: mode !== "full" && !forceRebuild,
            force_rebuild: forceRebuild || undefined,
            mode,
          },
          { headers },
        );
        // Fetch real progress first (may update stats)
        await fetchProgress();
        // Optimistic update: force status to running regardless of what GET returned
        setProgress((prev) =>
          prev
            ? {
                ...prev,
                status: "running" as const,
                lastError: undefined,
                updatedAt: Date.now(),
              }
            : ({
                status: "running" as const,
                mode: mode as "auto" | "full" | "incremental",
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
        );
        // Start polling with grace period of 5 (10 seconds)
        startPolling("running", 5);
      } catch (err) {
        if (err instanceof ApiError) {
          setError(err.message);
        } else {
          setError(err instanceof Error ? err.message : "Unknown error");
        }
      } finally {
        setLoading(false);
      }
    },
    [headers, fetchProgress, startPolling],
  );

  const inspectRoots = useCallback(
    async (folderIds: number[]): Promise<InspectRootsResponse | null> => {
      setInspectLoading(true);
      setInspectError(null);
      try {
        const result = await apiPost<InspectRootsResponse>(
          "/api/v1/admin/roots/inspect",
          {
            folder_ids: folderIds,
          },
          { headers },
        );
        const parsed = InspectRootsResponseSchema.parse(result);
        if (parsed.items.length > 0) {
          setProgress((prev) => {
            const now = Date.now();
            const base: SyncProgress =
              prev ??
              ({
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
              } satisfies SyncProgress);

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
        if (err instanceof ApiError) {
          setInspectError(err.message);
        } else {
          setInspectError(err instanceof Error ? err.message : "Unknown error");
        }
        return null;
      } finally {
        setInspectLoading(false);
      }
    },
    [headers],
  );

  const cancelSync = useCallback(async () => {
    setLoading(true);
    try {
      await apiDelete("/api/v1/admin/sync", { headers });
      await fetchProgress();
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError(err instanceof Error ? err.message : "Unknown error");
      }
    } finally {
      setLoading(false);
    }
  }, [headers, fetchProgress]);

  // Initial fetch — skip when no auth headers yet.
  const hasAuth = "X-API-Key" in headers;
  useEffect(() => {
    if (!hasAuth) return;

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
  };
}
