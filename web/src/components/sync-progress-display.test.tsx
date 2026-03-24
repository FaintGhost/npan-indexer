import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SyncProgressDisplay } from "./sync-progress-display";
import type { SyncProgress } from "@/lib/sync-schemas";

const baseProgress: SyncProgress = {
  status: "running",
  startedAt: 1700000000,
  updatedAt: 1700000500,
  roots: [100, 200],
  rootNames: {},
  completedRoots: [100],
  activeRoot: 200,
  mode: "full",
  aggregateStats: {
    foldersVisited: 50,
    filesIndexed: 300,
    filesDiscovered: 300,
    skippedFiles: 0,
    pagesFetched: 60,
    failedRequests: 2,
    startedAt: 1700000000,
    endedAt: 0,
  },
  rootProgress: {},
  lastError: "",
};

describe("SyncProgressDisplay", () => {
  it("shows running status", () => {
    render(<SyncProgressDisplay progress={baseProgress} />);
    expect(screen.getByText("运行中")).toBeInTheDocument();
  });

  it("shows roots progress", () => {
    render(<SyncProgressDisplay progress={baseProgress} />);
    expect(screen.getByText(/1.*\/.*2/)).toBeInTheDocument();
  });

  it("shows aggregate stats", () => {
    render(<SyncProgressDisplay progress={baseProgress} />);
    expect(screen.getByText("已索引文件").closest("div")).toHaveTextContent(
      "300",
    ); // filesIndexed
    expect(screen.getByText("已抓取页").closest("div")).toHaveTextContent(
      "60",
    ); // pagesFetched
  });

  it("shows done status", () => {
    render(
      <SyncProgressDisplay progress={{ ...baseProgress, status: "done" }} />,
    );
    expect(screen.getByText("已完成")).toBeInTheDocument();
  });

  it("shows error status with lastError", () => {
    render(
      <SyncProgressDisplay
        progress={{ ...baseProgress, status: "error", lastError: "网络超时" }}
      />,
    );
    expect(screen.getByText("出错")).toBeInTheDocument();
    expect(screen.getByText("网络超时")).toBeInTheDocument();
  });

  it("shows cancelled status", () => {
    render(
      <SyncProgressDisplay
        progress={{ ...baseProgress, status: "cancelled" }}
      />,
    );
    expect(screen.getByText("已取消")).toBeInTheDocument();
  });

  it("shows failed requests count when > 0", () => {
    render(<SyncProgressDisplay progress={baseProgress} />);
    expect(screen.getByText("失败请求")).toBeInTheDocument();
    const failedCard = screen.getByText("失败请求").closest("div")!;
    expect(failedCard).toHaveTextContent("2");
  });

  it("shows filesDiscovered stat", () => {
    const progress = {
      ...baseProgress,
      aggregateStats: { ...baseProgress.aggregateStats, filesDiscovered: 50 },
    };
    render(<SyncProgressDisplay progress={progress} />);
    expect(screen.getByText("已发现")).toBeInTheDocument();
  });

  it("shows verification success", () => {
    const progress: SyncProgress = {
      ...baseProgress,
      status: "done",
      verification: {
        meiliDocCount: 120,
        crawledDocCount: 120,
        discoveredDocCount: 120,
        skippedCount: 0,
        verified: true,
        warnings: [],
      },
    };
    render(<SyncProgressDisplay progress={progress} />);
    expect(screen.getByText("验证通过")).toBeInTheDocument();
  });

  it("shows verification warnings", () => {
    const progress: SyncProgress = {
      ...baseProgress,
      status: "done",
      verification: {
        meiliDocCount: 110,
        crawledDocCount: 120,
        discoveredDocCount: 120,
        skippedCount: 0,
        verified: false,
        warnings: ["索引文档数(110) < 爬取写入数(120)"],
      },
    };
    render(<SyncProgressDisplay progress={progress} />);
    expect(
      screen.getByText("索引文档数(110) < 爬取写入数(120)"),
    ).toBeInTheDocument();
  });

  it("hides verification when null", () => {
    render(<SyncProgressDisplay progress={baseProgress} />);
    expect(screen.queryByText("验证通过")).not.toBeInTheDocument();
  });

  it("renders incremental mode stats", () => {
    const incrementalProgress: SyncProgress = {
      ...baseProgress,
      mode: "incremental",
      incrementalStats: {
        changesFetched: 42,
        upserted: 30,
        deleted: 5,
        skippedUpserts: 3,
        skippedDeletes: 1,
        cursorBefore: 100,
        cursorAfter: 200,
      },
    };
    render(<SyncProgressDisplay progress={incrementalProgress} />);
    expect(screen.getByText("变更").closest("div")).toHaveTextContent("42");
    expect(screen.getByText("写入").closest("div")).toHaveTextContent("30");
    expect(screen.getByText("删除").closest("div")).toHaveTextContent("5");
    expect(screen.getByText("跳过写入").closest("div")).toHaveTextContent("3");
    expect(screen.getByText("跳过删除").closest("div")).toHaveTextContent("1");
  });

  it("shows root estimate and actual stats when estimate is present", async () => {
    const progress: SyncProgress = {
      ...baseProgress,
      rootNames: { 100: "PIXELHUE" },
      rootProgress: {
        "100": {
          rootFolderId: 100,
          status: "done",
          estimatedTotalDocs: 4152,
          stats: {
            foldersVisited: 119,
            filesIndexed: 393,
            filesDiscovered: 393,
            skippedFiles: 0,
            pagesFetched: 119,
            failedRequests: 0,
            startedAt: 0,
            endedAt: 0,
          },
          updatedAt: 0,
        },
      },
      roots: [100],
      completedRoots: [100],
    };

    render(<SyncProgressDisplay progress={progress} />);
    await userEvent.click(screen.getByRole("button", { name: /展开/i }));

    expect(screen.getByText(/官网总计 4,152/)).toBeInTheDocument();
    expect(screen.getByText(/索引统计 512/)).toBeInTheDocument();
  });

  it("renders catalogRootProgress when present", async () => {
    const progress: SyncProgress = {
      ...baseProgress,
      roots: [100],
      completedRoots: [100],
      rootProgress: {},
      catalogRoots: [100],
      catalogRootNames: { 100: "PIXELHUE" },
      catalogRootProgress: {
        "100": {
          rootFolderId: 100,
          status: "done",
          itemCount: 10,
          estimatedTotalDocs: 10,
          stats: {
            foldersVisited: 1,
            filesIndexed: 9,
            filesDiscovered: 9,
            skippedFiles: 0,
            pagesFetched: 1,
            failedRequests: 0,
            startedAt: 0,
            endedAt: 0,
          },
          updatedAt: 0,
        },
      },
    };

    render(<SyncProgressDisplay progress={progress} />);
    await userEvent.click(screen.getByRole("button", { name: /展开/i }));
    expect(screen.getByText(/PIXELHUE/)).toBeInTheDocument();
    expect(screen.getByText(/官网总计 10/)).toBeInTheDocument();
    expect(screen.getByText(/索引统计 10/)).toBeInTheDocument();
  });

  it("supports row toggle selection", async () => {
    const progress: SyncProgress = {
      ...baseProgress,
      roots: [100],
      completedRoots: [100],
      rootProgress: {
        "100": {
          rootFolderId: 100,
          status: "done",
          estimatedTotalDocs: 10,
          stats: {
            foldersVisited: 1,
            filesIndexed: 9,
            filesDiscovered: 9,
            skippedFiles: 0,
            pagesFetched: 1,
            failedRequests: 0,
            startedAt: 0,
            endedAt: 0,
          },
          updatedAt: 0,
        },
      },
    };

    const onToggleRoot = vi.fn();
    render(
      <SyncProgressDisplay
        progress={progress}
        rootSelection={{ selectedRootIds: [100], onToggleRoot }}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name: /展开/i }));
    await userEvent.click(screen.getByRole("switch", { name: /选择根目录 100/i }));
    expect(onToggleRoot).toHaveBeenCalledWith(100);
  });

  it("prefers timestamp sidecar fields for elapsed time display", () => {
    const progress = {
      ...baseProgress,
      status: "done",
      startedAt: 100,
      startedAtTs: { seconds: 1700000000, nanos: 0 },
      aggregateStats: {
        ...baseProgress.aggregateStats,
        endedAt: 101,
        endedAtTs: { seconds: 1700000005, nanos: 0 },
      },
    } as unknown as SyncProgress;

    render(<SyncProgressDisplay progress={progress} />);
    expect(screen.getByText(/用时 5s/)).toBeInTheDocument();
  });
});
