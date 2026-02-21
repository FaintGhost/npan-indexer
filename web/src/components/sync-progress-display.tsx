import { useState, useEffect } from 'react'
import type { SyncProgress } from '@/lib/sync-schemas'

interface SyncProgressDisplayProps {
  progress: SyncProgress
}

const statusConfig: Record<string, { label: string; color: string; badgeBg: string }> = {
  idle: { label: '空闲', color: 'text-slate-600', badgeBg: 'bg-slate-100' },
  running: { label: '运行中', color: 'text-blue-600', badgeBg: 'bg-blue-100' },
  done: { label: '已完成', color: 'text-emerald-600', badgeBg: 'bg-emerald-100' },
  error: { label: '出错', color: 'text-rose-600', badgeBg: 'bg-rose-100' },
  cancelled: { label: '已取消', color: 'text-slate-500', badgeBg: 'bg-slate-100' },
}

const defaultConfig = { label: '空闲', color: 'text-slate-600', badgeBg: 'bg-slate-100' } as const

export function SyncProgressDisplay({ progress }: SyncProgressDisplayProps) {
  const config = statusConfig[progress.status] ?? defaultConfig
  const stats = progress.aggregateStats
  const isRunning = progress.status === 'running'

  const rootsDone = progress.completedRoots.length
  const rootsTotal = progress.roots.length
  const rootPercent = rootsTotal > 0 ? Math.round((rootsDone / rootsTotal) * 100) : 0

  return (
    <div className="space-y-4">
      {/* Header: status + elapsed time */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className={`inline-flex items-center gap-1.5 rounded-lg px-3 py-1 text-sm font-medium ${config.badgeBg} ${config.color}`}>
            {isRunning && (
              <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-blue-500" />
            )}
            {config.label}
          </span>
        </div>
        {progress.startedAt > 0 && (
          <ElapsedTime
            startedAt={progress.startedAt}
            endedAt={stats.endedAt}
            isRunning={isRunning}
          />
        )}
      </div>

      {/* Root progress bar */}
      {rootsTotal > 0 && (
        <div className="rounded-xl border border-slate-200 bg-white p-4">
          <div className="flex items-center justify-between text-sm">
            <span className="font-medium text-slate-700">根目录进度</span>
            <span className="tabular-nums text-slate-500">
              {rootsDone} / {rootsTotal}
              <span className="ml-1 text-slate-400">({rootPercent}%)</span>
            </span>
          </div>
          <div className="mt-2 h-2 overflow-hidden rounded-full bg-slate-100">
            <div
              className={`h-full rounded-full transition-all duration-500 ease-out ${
                isRunning ? 'bg-blue-500' : progress.status === 'done' ? 'bg-emerald-500' : 'bg-slate-400'
              }`}
              style={{ width: `${rootPercent}%` }}
            />
          </div>
          {progress.activeRoot != null && isRunning && (
            <p className="mt-2 text-xs text-slate-400">
              当前根目录: <span className="font-mono">{progress.rootNames?.[String(progress.activeRoot)] || progress.activeRoot}</span>
            </p>
          )}
        </div>
      )}

      {/* Stats grid */}
      <div className="grid grid-cols-2 gap-3">
        <StatCard label="已索引文件" value={stats.filesIndexed} />
        <StatCard label="已抓取页" value={stats.pagesFetched} />
        <StatCard label="已访问文件夹" value={stats.foldersVisited} />
        {(stats.filesDiscovered ?? 0) > 0 && (
          <StatCard label="已发现" value={stats.filesDiscovered ?? 0} />
        )}
        {(stats.skippedFiles ?? 0) > 0 && (
          <StatCard label="已跳过" value={stats.skippedFiles ?? 0} skipped />
        )}
        {stats.failedRequests > 0 && (
          <StatCard label="失败请求" value={stats.failedRequests} error />
        )}
      </div>

      {/* Verification result */}
      {progress.verification != null && (
        progress.verification.warnings == null || progress.verification.warnings.length === 0 ? (
          <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-3">
            <p className="text-sm font-medium text-emerald-700">验证通过</p>
            <p className="mt-1 text-xs text-emerald-600">
              MeiliSearch 文档数: {progress.verification.meiliDocCount.toLocaleString()}
            </p>
          </div>
        ) : (
          <div className="rounded-xl border border-amber-200 bg-amber-50 p-3">
            <p className="text-sm font-medium text-amber-800">验证警告</p>
            <ul className="mt-1 list-disc pl-4">
              {progress.verification.warnings.map((w, i) => (
                <li key={i} className="text-xs text-amber-800">{w}</li>
              ))}
            </ul>
          </div>
        )
      )}

      {/* Per-root details (collapsed by default, expandable) */}
      {rootsTotal > 0 && <RootDetails progress={progress} />}

      {/* Error */}
      {progress.lastError && (
        <div className="rounded-xl border border-rose-200 bg-rose-50 p-3">
          <p className="text-sm text-rose-600">{progress.lastError}</p>
        </div>
      )}
    </div>
  )
}

function ElapsedTime({
  startedAt,
  endedAt,
  isRunning,
}: {
  startedAt: number
  endedAt: number
  isRunning: boolean
}) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (!isRunning) return
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [isRunning])

  const end = isRunning ? now : endedAt > 0 ? endedAt : Date.now()
  // startedAt might be in seconds or milliseconds
  const startMs = startedAt < 1e12 ? startedAt * 1000 : startedAt
  const endMs = end < 1e12 ? end * 1000 : end
  const elapsed = Math.max(0, Math.floor((endMs - startMs) / 1000))

  const hours = Math.floor(elapsed / 3600)
  const minutes = Math.floor((elapsed % 3600) / 60)
  const seconds = elapsed % 60

  const parts: string[] = []
  if (hours > 0) parts.push(`${hours}h`)
  if (minutes > 0 || hours > 0) parts.push(`${minutes}m`)
  parts.push(`${seconds}s`)

  return (
    <span className="tabular-nums text-sm text-slate-400">
      {isRunning ? '已耗时 ' : '用时 '}{parts.join(' ')}
    </span>
  )
}

function RootDetails({ progress }: { progress: SyncProgress }) {
  const [expanded, setExpanded] = useState(false)
  const entries = Object.entries(progress.rootProgress)
  if (entries.length === 0) return null

  return (
    <div className="rounded-xl border border-slate-200 bg-white">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between p-4 text-sm font-medium text-slate-700 hover:bg-slate-50"
      >
        <span>根目录详情 ({entries.length})</span>
        <span className="text-slate-400">{expanded ? '收起' : '展开'}</span>
      </button>
      {expanded && (
        <div className="border-t border-slate-100 divide-y divide-slate-100">
          {entries.map(([key, root]) => {
            const rootStatus = root.status === 'done'
              ? 'text-emerald-600'
              : root.status === 'running'
                ? 'text-blue-600'
                : root.status === 'error'
                  ? 'text-rose-600'
                  : 'text-slate-500'

            const rootName = progress.rootNames?.[key]

            return (
              <div key={key} className="px-4 py-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-slate-700">
                    {rootName ? (
                      <><span className="font-medium">{rootName}</span> <span className="font-mono text-xs text-slate-400">({root.rootFolderId})</span></>
                    ) : (
                      <span className="font-mono">{root.rootFolderId}</span>
                    )}
                  </span>
                  <span className={`text-xs font-medium ${rootStatus}`}>{root.status}</span>
                </div>
                <div className="mt-1 flex gap-4 text-xs text-slate-400">
                  <span>{root.stats.filesIndexed.toLocaleString()} 文件</span>
                  <span>{root.stats.foldersVisited.toLocaleString()} 文件夹</span>
                  <span>{root.stats.pagesFetched.toLocaleString()} 页</span>
                  {root.stats.failedRequests > 0 && (
                    <span className="text-rose-400">{root.stats.failedRequests} 失败</span>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

function StatCard({
  label,
  value,
  error = false,
  skipped = false,
}: {
  label: string
  value: number
  error?: boolean
  skipped?: boolean
}) {
  const bgClass = skipped ? 'bg-rose-50' : 'bg-white'
  const valueClass = error || skipped ? 'text-rose-600' : 'text-slate-900'
  return (
    <div className={`rounded-xl border border-slate-200 ${bgClass} p-3`}>
      <p className="text-xs text-slate-500">{label}</p>
      <p className={`mt-1 text-xl font-semibold tabular-nums ${valueClass}`}>
        {value.toLocaleString()}
      </p>
    </div>
  )
}
