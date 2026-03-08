import { useMemo } from 'react'
import { useCurrentRefinements, useRefinementList } from 'react-instantsearch'
import {
  DEFAULT_FILTER,
  SEARCH_FILTER_OPTIONS,
  getSearchFilterFromRefinement,
  type SearchFilter,
} from '@/lib/file-category'

function getSelectedFilter(items: ReturnType<typeof useCurrentRefinements>['items']): SearchFilter {
  const fileCategoryGroup = items.find((item) => item.attribute === 'file_category')
  const refinement = fileCategoryGroup?.refinements[0]
  return getSearchFilterFromRefinement(
    typeof refinement?.value === 'string' ? refinement.value : undefined,
  )
}

export function SearchFilters() {
  const { refine } = useRefinementList({
    attribute: 'file_category',
    operator: 'or',
    limit: SEARCH_FILTER_OPTIONS.length - 1,
  })
  const { items } = useCurrentRefinements({
    includedAttributes: ['file_category'],
  })

  const activeFilter = useMemo(() => getSelectedFilter(items), [items])

  const handleFilterChange = (nextFilter: SearchFilter) => {
    if (nextFilter === activeFilter) {
      return
    }

    if (activeFilter !== DEFAULT_FILTER) {
      refine(activeFilter)
    }
    if (nextFilter !== DEFAULT_FILTER) {
      refine(nextFilter)
    }
  }

  const handleFilterKeyDown = (
    event: React.KeyboardEvent<HTMLButtonElement>,
    current: SearchFilter,
  ) => {
    if (!['ArrowRight', 'ArrowDown', 'ArrowLeft', 'ArrowUp'].includes(event.key)) {
      return
    }

    event.preventDefault()
    const currentIndex = SEARCH_FILTER_OPTIONS.findIndex((option) => option.value === current)
    if (currentIndex < 0) {
      return
    }

    const isForward = event.key === 'ArrowRight' || event.key === 'ArrowDown'
    const nextIndex = isForward
      ? (currentIndex + 1) % SEARCH_FILTER_OPTIONS.length
      : (currentIndex - 1 + SEARCH_FILTER_OPTIONS.length) % SEARCH_FILTER_OPTIONS.length
    const nextFilter = SEARCH_FILTER_OPTIONS[nextIndex]?.value

    if (nextFilter) {
      handleFilterChange(nextFilter)
    }
  }

  return (
    <div className="mt-4" role="radiogroup" aria-label="结果筛选">
      <div className="flex flex-wrap gap-2.5">
        {SEARCH_FILTER_OPTIONS.map((option) => {
          const checked = activeFilter === option.value

          return (
            <button
              key={option.value}
              type="button"
              role="radio"
              aria-checked={checked}
              tabIndex={checked ? 0 : -1}
              onClick={() => handleFilterChange(option.value)}
              onKeyDown={(event) => handleFilterKeyDown(event, option.value)}
              className={checked
                ? 'rounded-xl border border-blue-200 bg-blue-50 px-3 py-1.5 text-xs font-semibold text-blue-800 shadow-sm'
                : 'rounded-xl border border-slate-200 bg-white/95 px-3 py-1.5 text-xs font-medium text-slate-600 hover:border-slate-300'}
            >
              {option.label}
            </button>
          )
        })}
      </div>
    </div>
  )
}
