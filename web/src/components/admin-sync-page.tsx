import { useState } from 'react'
import { useAdminAuth } from '@/hooks/use-admin-auth'
import { useSyncProgress } from '@/hooks/use-sync-progress'
import { ApiKeyDialog } from '@/components/api-key-dialog'
import { SyncProgressDisplay } from '@/components/sync-progress-display'

const SYNC_MODES = [
  { value: 'auto', label: '自适应', description: '有游标走增量，否则全量' },
  { value: 'full', label: '全量', description: '重新爬取所有目录' },
  { value: 'incremental', label: '增量', description: '仅同步最近变更' },
] as const

export function AdminSyncPage() {
  const auth = useAdminAuth()
  const sync = useSyncProgress(auth.getHeaders())
  const [message, setMessage] = useState<string | null>(null)
  const [mode, setMode] = useState<string>('auto')

  const isRunning = sync.progress?.status === 'running'
  const isBusy = sync.loading || isRunning

  const handleStartSync = async () => {
    setMessage(null)
    await sync.startSync([], mode)
    if (!sync.error) {
      setMessage('同步任务已启动')
      setTimeout(() => setMessage(null), 4000)
    }
  }

  const handleCancelSync = async () => {
    if (!window.confirm('确认取消同步？')) return
    setMessage(null)
    await sync.cancelSync()
    if (!sync.error) {
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
                    ? 'bg-white text-slate-900 shadow-sm'
                    : 'text-slate-500 hover:text-slate-700'
                } disabled:cursor-not-allowed disabled:opacity-60`}
                title={m.description}
              >
                {m.label}
              </button>
            ))}
          </div>
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
              ? '启动中...'
              : isRunning
                ? '同步进行中'
                : '启动同步'}
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

      {!sync.initialLoading && sync.progress && <SyncProgressDisplay progress={sync.progress} />}

      {!sync.initialLoading && !sync.progress && !sync.loading && !sync.error && (
        <div className="py-12 text-center">
          <p className="text-sm text-slate-400">暂无同步记录</p>
        </div>
      )}
    </div>
  )
}
