export function InitialState() {
  return (
    <div className="rounded-3xl border border-dashed border-slate-300 bg-slate-50/50 px-6 py-16 text-center">
      <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-blue-100/50 text-blue-500">
        <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
      </div>
      <h3 className="text-[15px] font-medium text-slate-900">等待探索</h3>
      <p className="mt-2 text-sm text-slate-500">
        输入你想查找的文件名、格式或相关关键词，
        <br className="hidden sm:block" />
        我们将为你检索全库资源。
      </p>
    </div>
  )
}

export function NoResultsState() {
  return (
    <div className="rounded-3xl border border-dashed border-slate-300 bg-slate-50/50 px-6 py-16 text-center">
      <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-slate-200/50 text-slate-500">
        <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
          <line x1="9" x2="15" y1="13" y2="13" />
          <line x1="12" x2="12" y1="10" y2="16" />
        </svg>
      </div>
      <h3 className="text-[15px] font-medium text-slate-900">未找到相关文件</h3>
      <p className="mt-2 text-sm text-slate-500">
        抱歉，没有找到匹配的内容。建议尝试更简短或更准确的关键词。
      </p>
    </div>
  )
}

export function ErrorState() {
  return (
    <div className="rounded-3xl border border-dashed border-rose-200 bg-rose-50/50 px-6 py-16 text-center">
      <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-rose-100/50 text-rose-500">
        <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="12" cy="12" r="10" />
          <line x1="12" x2="12" y1="8" y2="12" />
          <line x1="12" x2="12.01" y1="16" y2="16" />
        </svg>
      </div>
      <h3 className="text-[15px] font-medium text-slate-900">加载出错了</h3>
      <p className="mt-2 text-sm text-slate-500">
        网络请求似乎遇到了问题，请稍后重试或联系后端程序。
      </p>
    </div>
  )
}
