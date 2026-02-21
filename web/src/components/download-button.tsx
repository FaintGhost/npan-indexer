type DownloadStatus = 'idle' | 'loading' | 'success' | 'error'

interface DownloadButtonProps {
  status: DownloadStatus
  onClick: () => void
}

export function DownloadButton({ status, onClick }: DownloadButtonProps) {
  const isDisabled = status === 'loading' || status === 'success'

  return (
    <button
      type="button"
      onClick={onClick}
      disabled={isDisabled}
      className="relative flex h-10 w-full sm:w-auto min-w-[96px] shrink-0 items-center justify-center rounded-xl bg-slate-900 px-4 text-sm font-medium text-white transition-all hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-700 disabled:opacity-90"
    >
      {status === 'idle' && (
        <span className="flex items-center gap-1.5">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" x2="12" y1="15" y2="3" />
          </svg>
          下载
        </span>
      )}
      {status === 'loading' && (
        <span className="flex items-center gap-1.5">
          <svg className="animate-spin text-white/80" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 12a9 9 0 1 1-6.219-8.56" />
          </svg>
          获取中
        </span>
      )}
      {status === 'success' && (
        <span className="flex items-center gap-1.5 text-emerald-400">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="20 6 9 17 4 12" />
          </svg>
          成功
        </span>
      )}
      {status === 'error' && (
        <span className="flex items-center gap-1.5 text-rose-400">重试</span>
      )}
    </button>
  )
}
