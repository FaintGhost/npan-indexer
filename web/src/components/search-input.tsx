import { isMac } from '@/hooks/use-hotkey'

interface SearchInputProps {
  value: string
  onChange: (value: string) => void
  onSubmit: () => void
  onClear: () => void
  ref?: React.Ref<HTMLInputElement>
}

export function SearchInput({ value, onChange, onSubmit, onClear, ref }: SearchInputProps) {
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      onSubmit()
    }
  }

  return (
    <div className="relative flex w-full items-center">
      {/* Search icon */}
      <div className="pointer-events-none absolute left-4 text-slate-400">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.3-4.3" />
        </svg>
      </div>

      <input
        ref={ref}
        type="text"
        role="searchbox"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="输入文件名关键词，例如：MX40、固件、安装包"
        autoComplete="off"
        className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-12 pr-16 text-[15px] text-slate-800 outline-none ring-blue-100 transition focus:border-blue-300 focus:ring-4"
      />

      <div className="absolute right-3 flex items-center gap-2">
        {value ? (
          <button
            type="button"
            onClick={onClear}
            aria-label="清空搜索"
            className="rounded-full p-1.5 text-slate-400 transition hover:bg-slate-100 hover:text-slate-600 focus:outline-none"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M18 6 6 18" />
              <path d="m6 6 12 12" />
            </svg>
          </button>
        ) : (
          <kbd className="hidden items-center justify-center rounded border border-slate-200 bg-slate-50 px-1.5 py-0.5 text-[10px] font-medium text-slate-400 sm:inline-flex">
            {isMac() ? '⌘K' : 'Ctrl K'}
          </kbd>
        )}
      </div>
    </div>
  )
}
