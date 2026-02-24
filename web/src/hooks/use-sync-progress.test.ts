import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "../tests/mocks/server";
import { useSyncProgress } from "./use-sync-progress";
import type { InspectRootsResponse } from "@/lib/sync-schemas";

function assertRecord(value: unknown): asserts value is Record<string, unknown> {
  if (typeof value !== "object" || value === null) {
    throw new Error("expected payload to be an object");
  }
}

function getRecord(value: unknown): Record<string, unknown> {
  assertRecord(value);
  return value;
}

function requireValue<T>(
  value: T | null | undefined,
  message: string,
): NonNullable<T> {
  if (value == null) {
    throw new Error(message);
  }
  return value;
}

const validProgress = {
  status: "idle",
  startedAt: 0,
  updatedAt: 0,
  roots: [],
  completedRoots: [],
  aggregateStats: {
    foldersVisited: 0,
    filesIndexed: 0,
    filesDiscovered: 0,
    skippedFiles: 0,
    pagesFetched: 0,
    failedRequests: 0,
    startedAt: 0,
    endedAt: 0,
  },
  rootProgress: {},
};

describe("useSyncProgress", () => {
  const headers = { "X-API-Key": "test-key" };

  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("fetches initial progress", async () => {
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json({
          ...validProgress,
          status: "done",
          roots: [100, 200],
          completedRoots: [100, 200],
          aggregateStats: {
            ...validProgress.aggregateStats,
            filesIndexed: 500,
          },
        });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
      expect(result.current.progress?.status).toBe("done");
      expect(result.current.progress?.aggregateStats.filesIndexed).toBe(500);
    });
  });

  it("prefers timestamp sidecar fields when present", async () => {
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json({
          ...validProgress,
          startedAt: 1,
          updatedAt: 2,
          startedAtTs: { seconds: 1700000000, nanos: 123000000 },
          updatedAtTs: { seconds: 1700000005, nanos: 456000000 },
          aggregateStats: {
            ...validProgress.aggregateStats,
            startedAt: 3,
            endedAt: 4,
            startedAtTs: { seconds: 1700000001, nanos: 0 },
            endedAtTs: { seconds: 1700000002, nanos: 250000000 },
          },
        });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
      expect(result.current.progress?.startedAt).toBe(1700000000123);
      expect(result.current.progress?.updatedAt).toBe(1700000005456);
      expect(result.current.progress?.aggregateStats.startedAt).toBe(
        1700000001000,
      );
      expect(result.current.progress?.aggregateStats.endedAt).toBe(
        1700000002250,
      );
    });
  });

  it("starts sync", async () => {
    let postCalled = false;
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json(validProgress);
      }),
      http.post("/api/v1/admin/sync", () => {
        postCalled = true;
        return HttpResponse.json({ message: "Sync started" });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
    });

    await act(async () => {
      await result.current.startSync([100, 200]);
    });

    expect(postCalled).toBe(true);
  });

  it("sends resume_progress=false when mode is full", async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json(validProgress);
      }),
      http.post("/api/v1/admin/sync", async ({ request }) => {
        const body: unknown = await request.json();
        assertRecord(body);
        capturedBody = body;
        return HttpResponse.json({ message: "Sync started" });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
    });

    await act(async () => {
      await result.current.startSync([], "full");
    });

    expect(capturedBody).not.toBeNull();
    const payload = getRecord(capturedBody);
    expect(payload.resume_progress).toBe(false);
    expect(payload.mode).toBe("full");
  });

  it("sends resume_progress=true when mode is auto", async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json(validProgress);
      }),
      http.post("/api/v1/admin/sync", async ({ request }) => {
        const body: unknown = await request.json();
        assertRecord(body);
        capturedBody = body;
        return HttpResponse.json({ message: "Sync started" });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
    });

    await act(async () => {
      await result.current.startSync([], "auto");
    });

    expect(capturedBody).not.toBeNull();
    const payload = getRecord(capturedBody);
    expect(payload.resume_progress).toBe(true);
    expect(payload.mode).toBe("auto");
  });

  it("sends force_rebuild when forceRebuild is true", async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json(validProgress);
      }),
      http.post("/api/v1/admin/sync", async ({ request }) => {
        const body: unknown = await request.json();
        assertRecord(body);
        capturedBody = body;
        return HttpResponse.json({ message: "Sync started" });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
    });

    await act(async () => {
      await result.current.startSync([], "full", true);
    });

    expect(capturedBody).not.toBeNull();
    const payload = getRecord(capturedBody);
    expect(payload.force_rebuild).toBe(true);
    expect(payload.resume_progress).toBe(false);
  });

  it("sends include_departments=false when root_folder_ids is provided", async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json(validProgress);
      }),
      http.post("/api/v1/admin/sync", async ({ request }) => {
        const body: unknown = await request.json();
        assertRecord(body);
        capturedBody = body;
        return HttpResponse.json({ message: "Sync started" });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
    });

    await act(async () => {
      await result.current.startSync([123456], "full");
    });

    expect(capturedBody).not.toBeNull();
    const payload = getRecord(capturedBody);
    expect(payload.root_folder_ids).toEqual([123456]);
    expect(payload.include_departments).toBe(false);
    expect(payload.preserve_root_catalog).toBe(true);
  });

  it("inspects roots without starting sync", async () => {
    let inspectCalled = false;
    let syncPostCalled = false;
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json(validProgress);
      }),
      http.post("/api/v1/admin/roots/inspect", async ({ request }) => {
        inspectCalled = true;
        const body: unknown = await request.json();
        assertRecord(body);
        expect(body.folder_ids).toEqual([1001, 1002]);
        return HttpResponse.json({
          items: [
            {
              folder_id: 1001,
              name: "A",
              item_count: 10,
              estimated_total_docs: 11,
            },
          ],
          errors: [{ folder_id: 1002, message: "获取目录信息失败" }],
        });
      }),
      http.post("/api/v1/admin/sync", () => {
        syncPostCalled = true;
        return HttpResponse.json({ message: "Sync started" });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
    });

    let inspectResult: InspectRootsResponse | null = null;
    await act(async () => {
      const response = await result.current.inspectRoots([1001, 1002]);
      if (!response) {
        throw new Error("expected inspect result");
      }
      inspectResult = response;
    });

    expect(inspectCalled).toBe(true);
    expect(syncPostCalled).toBe(false);
    const inspectResponse = requireValue<InspectRootsResponse>(
      inspectResult,
      "expected inspect result",
    );
    expect(inspectResponse.items).toHaveLength(1);
    expect(result.current.progress?.catalogRoots).toContain(1001);
  });

  it("omits force_rebuild when forceRebuild is false", async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json(validProgress);
      }),
      http.post("/api/v1/admin/sync", async ({ request }) => {
        const body: unknown = await request.json();
        assertRecord(body);
        capturedBody = body;
        return HttpResponse.json({ message: "Sync started" });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
    });

    await act(async () => {
      await result.current.startSync([], "full", false);
    });

    expect(capturedBody).not.toBeNull();
    const payload = getRecord(capturedBody);
    expect(payload.force_rebuild).toBeUndefined();
  });

  it("cancels sync", async () => {
    let cancelCalled = false;
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json({ ...validProgress, status: "running" });
      }),
      http.delete("/api/v1/admin/sync", () => {
        cancelCalled = true;
        return HttpResponse.json({ message: "Cancelled" });
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.progress).not.toBeNull();
    });

    await act(async () => {
      await result.current.cancelSync();
    });

    expect(cancelCalled).toBe(true);
  });

  it("sets error on failed request", async () => {
    server.use(
      http.get("/api/v1/admin/sync", () => {
        return HttpResponse.json(
          { code: "INTERNAL_ERROR", message: "Server error" },
          { status: 500 },
        );
      }),
    );

    const { result } = renderHook(() => useSyncProgress(headers));

    await waitFor(() => {
      expect(result.current.error).toBeTruthy();
    });
  });

  describe("startSync 后状态自动刷新", () => {
    const doneProgress = {
      ...validProgress,
      status: "done",
      startedAt: 1000,
      updatedAt: 2000,
      roots: [100],
      completedRoots: [100],
      aggregateStats: {
        ...validProgress.aggregateStats,
        filesIndexed: 42,
      },
    };

    it("startSync 后 progress 立即变为 running（乐观更新）", async () => {
      let getCalls = 0;
      server.use(
        http.get("/api/v1/admin/sync", () => {
          getCalls++;
          return HttpResponse.json(doneProgress);
        }),
        http.post("/api/v1/admin/sync", () => {
          return HttpResponse.json({ message: "ok" }, { status: 202 });
        }),
      );

      const { result } = renderHook(() => useSyncProgress(headers));

      // Wait for initial fetch
      await waitFor(() => {
        expect(result.current.progress).not.toBeNull();
      });

      // startSync should set progress.status to "running" optimistically,
      // even though GET returns "done" (old data)
      await act(async () => {
        await result.current.startSync([100], "auto");
      });

      expect(result.current.progress?.status).toBe("running");
    });

    it("startSync 后轮询不因旧数据停止（宽限期内）", async () => {
      let getCalls = 0;
      server.use(
        http.get("/api/v1/admin/sync", () => {
          getCalls++;
          return HttpResponse.json(doneProgress);
        }),
        http.post("/api/v1/admin/sync", () => {
          return HttpResponse.json({ message: "ok" }, { status: 202 });
        }),
      );

      const { result } = renderHook(() => useSyncProgress(headers));

      // Wait for initial fetch
      await waitFor(() => {
        expect(result.current.progress).not.toBeNull();
      });

      await act(async () => {
        await result.current.startSync([100], "auto");
      });

      // Record calls right after startSync (includes the fetchProgress in startSync)
      const callsAfterSync = getCalls;

      // Advance past one poll interval — polling should still be active
      // even though GET keeps returning "done" (grace period)
      await act(async () => {
        vi.advanceTimersByTime(2000);
      });

      const callsAfterFirstPoll = getCalls;
      expect(callsAfterFirstPoll).toBeGreaterThan(callsAfterSync);

      // Advance another poll interval — still within grace period
      await act(async () => {
        vi.advanceTimersByTime(2000);
      });

      const callsAfterSecondPoll = getCalls;
      expect(callsAfterSecondPoll).toBeGreaterThan(callsAfterFirstPoll);
    });

    it("宽限期结束后轮询正常停止", async () => {
      let getCalls = 0;
      server.use(
        http.get("/api/v1/admin/sync", () => {
          getCalls++;
          return HttpResponse.json(doneProgress);
        }),
        http.post("/api/v1/admin/sync", () => {
          return HttpResponse.json({ message: "ok" }, { status: 202 });
        }),
      );

      const { result } = renderHook(() => useSyncProgress(headers));

      // Wait for initial fetch
      await waitFor(() => {
        expect(result.current.progress).not.toBeNull();
      });

      await act(async () => {
        await result.current.startSync([100], "auto");
      });

      // Advance past the grace period (5 polls × 2000ms = 10000ms + extra)
      await act(async () => {
        vi.advanceTimersByTime(12000);
      });

      // Record the call count after grace period expires
      const callsAfterGrace = getCalls;

      // Advance another poll interval — polling should have stopped
      await act(async () => {
        vi.advanceTimersByTime(4000);
      });

      const callsAfterExtra = getCalls;
      expect(callsAfterExtra).toBe(callsAfterGrace);
    });
  });
});
