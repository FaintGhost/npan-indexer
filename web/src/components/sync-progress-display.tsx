import type { SyncProgress } from '@/lib/sync-schemas'

interface SyncProgressDisplayProps {
  progress: SyncProgress
}

const defaultConfig = { label: '空闲', color: 'bg-slate-100 text-slate-600' }

const statusConfig: Record<string, { label: string; color: string }> = {
  idle: defaultConfig,
  running: { label: '运行中', color: 'bg-blue-100 text-blue-600' },
  done: { label: '已完成', color: 'bg-emerald-100 text-emerald-600' },
  error: { label: '出错', color: 'bg-rose-100 text-rose-600' },
  cancelled: { label: '已取消', color: 'bg-slate-100 text-slate-500' },
}

export function SyncProgressDisplay({ progress }: SyncProgressDisplayProps) {
  const config = statusConfig[progress.status] ?? defaultConfig
  const stats = progress.aggregateStats

  return (
    <div className="space-y-4">
      {/* Status badge */}
      <div className="flex items-center gap-3">
        <span className={`inline-flex items-center rounded-lg px-3 py-1 text-sm font-medium ${config.color}`}>
          {config.label}
        </span>
      </div>

      {/* Roots progress */}
      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <h3 className="text-sm font-medium text-slate-700">根目录进度</h3>
        <p className="mt-1 text-2xl font-semibold text-slate-900">
          {progress.completedRoots.length} / {progress.roots.length}
        </p>
        {progress.activeRoot != null && (
          <p className="mt-1 text-sm text-slate-500">
            当前: {progress.activeRoot}
          </p>
        )}
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-2 gap-3">
        <StatCard label="已索引文件" value={stats.filesIndexed} />
        <StatCard label="已抓取页" value={stats.pagesFetched} />
        <StatCard label="已访问文件夹" value={stats.foldersVisited} />
        {stats.failedRequests > 0 && (
          <StatCard
            label="失败请求"
            value={stats.failedRequests}
            className="text-rose-600"
          />
        )}
      </div>

      {/* Error message */}
      {progress.lastError && (
        <div className="rounded-xl border border-rose-200 bg-rose-50 p-3">
          <p className="text-sm text-rose-600">{progress.lastError}</p>
        </div>
      )}
    </div>
  )
}

function StatCard({
  label,
  value,
  className = '',
}: {
  label: string
  value: number
  className?: string
}) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3">
      <p className="text-xs text-slate-500">{label}</p>
      <p className={`mt-1 text-xl font-semibold text-slate-900 ${className}`}>
        {value.toLocaleString()}
      </p>
    </div>
  )
}
