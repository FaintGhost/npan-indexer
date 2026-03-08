import { useMemo } from 'react'
import { useCurrentRefinements, useRefinementList } from 'react-instantsearch'
import {
  DEFAULT_FILTER,
  SEARCH_FILTER_OPTIONS,
  getSearchFilterFromRefinement,
  type SearchFilter,
} from '@/lib/file-category'

type FileCategoryRefinementItem = ReturnType<typeof useRefinementList>['items'][number]
type FileCategoryItemMap = Partial<Record<Exclude<SearchFilter, 'all'>, FileCategoryRefinementItem>>
type CurrentRefinementItems = ReturnType<typeof useCurrentRefinements>['items']

function getFilterFromItem(item: FileCategoryRefinementItem | undefined): SearchFilter {
  return getSearchFilterFromRefinement(typeof item?.label === 'string' ? item.label : undefined)
}

function getSelectedFilter(
  currentRefinementItems: CurrentRefinementItems,
  refinementItems: Array<FileCategoryRefinementItem>,
): SearchFilter {
  const activeItem = refinementItems.find((item) => item.isRefined && getFilterFromItem(item) !== DEFAULT_FILTER)
  if (activeItem) {
    return getFilterFromItem(activeItem)
  }

  const fileCategoryGroup = currentRefinementItems.find((item) => item.attribute === 'file_category')
  const refinement = fileCategoryGroup?.refinements[0]
  return getSearchFilterFromRefinement(
    typeof refinement?.value === 'string' ? refinement.value : undefined,
  )
}

function getFilterItemMap(items: Array<FileCategoryRefinementItem>): FileCategoryItemMap {
  return items.reduce<FileCategoryItemMap>((result, item) => {
    const filter = getFilterFromItem(item)
    if (filter !== DEFAULT_FILTER) {
      result[filter] = item
    }
    return result
  }, {})
}

function getRefinementToken(
  filter: Exclude<SearchFilter, 'all'>,
  filterItems: FileCategoryItemMap,
): string {
  const token = filterItems[filter]?.value
  return typeof token === 'string' && token.length > 0 ? token : filter
}

export function SearchFilters() {
  const { items: refinementItems, refine } = useRefinementList({
    attribute: 'file_category',
    operator: 'or',
    limit: SEARCH_FILTER_OPTIONS.length - 1,
  })
  const { items: currentRefinementItems } = useCurrentRefinements({
    includedAttributes: ['file_category'],
  })

  const activeFilter = useMemo(
    () => getSelectedFilter(currentRefinementItems, refinementItems),
    [currentRefinementItems, refinementItems],
  )
  const filterItems = useMemo(() => getFilterItemMap(refinementItems), [refinementItems])

  const handleFilterChange = (nextFilter: SearchFilter) => {
    if (nextFilter === activeFilter) {
      return
    }

    const activeToken = activeFilter === DEFAULT_FILTER
      ? undefined
      : getRefinementToken(activeFilter, filterItems)
    if (nextFilter === DEFAULT_FILTER) {
      if (activeToken) {
        refine(activeToken)
      }
      return
    }

    const nextToken = getRefinementToken(nextFilter, filterItems)
    if (activeToken) {
      refine(activeToken)
    }
    refine(nextToken)
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
