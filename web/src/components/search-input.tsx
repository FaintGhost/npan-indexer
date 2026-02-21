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
      onSubmit()
    }
  }

  return (
    <div className="relative">
      {/* Search icon */}
      <svg
        className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
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

      <input
        ref={ref}
        type="search"
        role="searchbox"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="搜索文件名..."
        className="w-full rounded-2xl border border-slate-200 bg-white py-4 pl-12 pr-12 text-base text-slate-900 outline-none transition-shadow focus:border-blue-300 focus:ring-4 focus:ring-blue-100"
        autoComplete="off"
      />

      {/* Clear button or keyboard shortcut badge */}
      {value ? (
        <button
          type="button"
          onClick={onClear}
          aria-label="清空搜索"
          className="absolute right-3 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <line x1="18" x2="6" y1="6" y2="18" />
            <line x1="6" x2="18" y1="6" y2="18" />
          </svg>
        </button>
      ) : (
        <kbd className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 rounded-lg border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs font-medium text-slate-400">
          {isMac() ? '⌘K' : 'Ctrl K'}
        </kbd>
      )}
    </div>
  )
}
