interface SkeletonCardProps {
  delay?: number
}

export function SkeletonCard({ delay = 0 }: SkeletonCardProps) {
  return (
    <div
      aria-hidden="true"
      className="animate-soft-pulse rounded-2xl border border-slate-200 bg-white px-4 py-4 sm:px-5"
      style={{ animationDelay: `${delay}ms` }}
    >
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 flex-1 items-start gap-4">
          <div className="h-12 w-12 shrink-0 rounded-xl bg-slate-100" />
          <div className="flex-1 space-y-3 pt-1.5">
            <div className="h-4 w-3/4 max-w-[280px] rounded-md bg-slate-200" />
            <div className="h-3 w-1/2 max-w-[200px] rounded-md bg-slate-100/80" />
          </div>
        </div>
        <div className="h-10 w-full shrink-0 rounded-xl bg-slate-100 sm:w-[96px]" />
      </div>
    </div>
  )
}
