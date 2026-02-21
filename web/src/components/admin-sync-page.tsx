import { useAdminAuth } from '@/hooks/use-admin-auth'
import { useSyncProgress } from '@/hooks/use-sync-progress'
import { ApiKeyDialog } from '@/components/api-key-dialog'
import { SyncProgressDisplay } from '@/components/sync-progress-display'

export function AdminSyncPage() {
  const auth = useAdminAuth()
  const sync = useSyncProgress(auth.getHeaders())

  const handleStartSync = async () => {
    // Use empty array to let server use defaults
    await sync.startSync([])
  }

  const handleCancelSync = async () => {
    if (window.confirm('确认取消同步？')) {
      await sync.cancelSync()
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

      {/* Action buttons */}
      <div className="mb-6 flex gap-3">
        <button
          type="button"
          onClick={handleStartSync}
          disabled={sync.loading || sync.progress?.status === 'running'}
          className="rounded-xl bg-slate-900 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-slate-800 disabled:opacity-60"
        >
          启动全量同步
        </button>

        {sync.progress?.status === 'running' && (
          <button
            type="button"
            onClick={handleCancelSync}
            disabled={sync.loading}
            className="rounded-xl border border-rose-200 bg-white px-5 py-2.5 text-sm font-medium text-rose-600 transition-colors hover:bg-rose-50 disabled:opacity-60"
          >
            取消同步
          </button>
        )}
      </div>

      {/* Error message */}
      {sync.error && (
        <div className="mb-4 rounded-xl border border-rose-200 bg-rose-50 p-3">
          <p className="text-sm text-rose-600">{sync.error}</p>
        </div>
      )}

      {/* Progress display */}
      {sync.progress && <SyncProgressDisplay progress={sync.progress} />}

      {!sync.progress && !sync.loading && !sync.error && (
        <div className="py-12 text-center">
          <p className="text-sm text-slate-400">暂无同步记录</p>
        </div>
      )}
    </div>
  )
}
