import { useState } from "react";
import { useAdminAuth } from "@/hooks/use-admin-auth";
import { useSyncProgress } from "@/hooks/use-sync-progress";
import { ApiKeyDialog } from "@/components/api-key-dialog";
import { SyncProgressDisplay } from "@/components/sync-progress-display";
import { ConfirmDialog } from "@/components/confirm-dialog";

const SYNC_MODES = [
  { value: "auto", label: "自适应", description: "有游标走增量，否则全量" },
  { value: "full", label: "全量", description: "重新爬取所有目录" },
  { value: "incremental", label: "增量", description: "仅同步最近变更" },
] as const;

export function AdminSyncPage() {
  const auth = useAdminAuth();
  const sync = useSyncProgress(auth.getHeaders());
  const [message, setMessage] = useState<string | null>(null);
  const [mode, setMode] = useState<string>("auto");
  const [forceRebuild, setForceRebuild] = useState(false);
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

  const isRunning = sync.progress?.status === "running";
  const isBusy = sync.loading || isRunning;

  const handleStartSync = async () => {
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
    await sync.startSync([], mode, forceRebuild);
    if (!sync.error) {
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
    await sync.cancelSync();
    if (!sync.error) {
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

      {/* Mode selector + Action buttons */}
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
                清空现有索引后重新爬取，重建期间搜索将无结果
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
            {sync.loading && !isRunning && (
              <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
            )}
            {isRunning && (
              <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-emerald-400" />
            )}
            {sync.loading && !isRunning
              ? "启动中..."
              : isRunning
                ? "同步进行中"
                : "启动同步"}
          </button>

          {isRunning && (
            <button
              type="button"
              onClick={handleCancelSync}
              disabled={sync.loading}
              className="rounded-xl border border-rose-200 bg-white px-5 py-2.5 text-sm font-medium text-rose-600 transition-colors hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-60"
            >
              取消同步
            </button>
          )}
        </div>
      </div>

      {/* Success message */}
      {message && (
        <div className="mb-4 rounded-xl border border-emerald-200 bg-emerald-50 p-3">
          <p className="text-sm text-emerald-700">{message}</p>
        </div>
      )}

      {/* Error message */}
      {sync.error && (
        <div className="mb-4 rounded-xl border border-rose-200 bg-rose-50 p-3">
          <p className="text-sm text-rose-600">{sync.error}</p>
        </div>
      )}

      {/* Progress display */}
      {sync.initialLoading && (
        <div className="space-y-4">
          <div className="h-8 w-24 animate-pulse rounded-lg bg-slate-200" />
          <div className="h-24 animate-pulse rounded-xl bg-slate-100" />
          <div className="grid grid-cols-2 gap-3">
            <div className="h-20 animate-pulse rounded-xl bg-slate-100" />
            <div className="h-20 animate-pulse rounded-xl bg-slate-100" />
          </div>
        </div>
      )}

      {!sync.initialLoading && sync.progress && (
        <SyncProgressDisplay progress={sync.progress} />
      )}

      {!sync.initialLoading &&
        !sync.progress &&
        !sync.loading &&
        !sync.error && (
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
