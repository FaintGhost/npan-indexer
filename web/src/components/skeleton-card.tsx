export function SkeletonCard() {
  return (
    <div
      aria-hidden="true"
      className="animate-pulse rounded-2xl border border-slate-200 bg-white px-4 py-4 sm:px-5"
    >
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-start gap-4 min-w-0 flex-1">
          <div className="h-12 w-12 shrink-0 rounded-xl bg-slate-100" />
          <div className="min-w-0 flex-1 space-y-2 pt-1">
            <div className="h-4 w-3/4 rounded bg-slate-100" />
            <div className="h-3 w-1/2 rounded bg-slate-100" />
          </div>
        </div>
        <div className="h-10 w-full rounded-xl bg-slate-100 sm:w-24" />
      </div>
    </div>
  )
}
