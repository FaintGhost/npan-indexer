export function InitialState() {
  return (
    <div className="flex flex-col items-center py-20 text-center">
      <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="28"
          height="28"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-slate-400"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
      </div>
      <h2 className="mt-5 text-lg font-semibold text-slate-700">等待探索</h2>
      <p className="mt-1 text-sm text-slate-400">输入关键词开始搜索文件</p>
    </div>
  )
}

export function NoResultsState() {
  return (
    <div className="flex flex-col items-center py-20 text-center">
      <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="28"
          height="28"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-slate-400"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
          <line x1="8" x2="14" y1="11" y2="11" />
        </svg>
      </div>
      <h2 className="mt-5 text-lg font-semibold text-slate-700">未找到相关文件</h2>
      <p className="mt-1 text-sm text-slate-400">试试其他关键词</p>
    </div>
  )
}

export function ErrorState() {
  return (
    <div className="flex flex-col items-center py-20 text-center">
      <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-rose-50">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="28"
          height="28"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-rose-400"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" x2="12" y1="8" y2="12" />
          <line x1="12" x2="12.01" y1="16" y2="16" />
        </svg>
      </div>
      <h2 className="mt-5 text-lg font-semibold text-rose-600">加载出错了</h2>
      <p className="mt-1 text-sm text-slate-400">请稍后重试</p>
    </div>
  )
}
