import type { StateMapping, UiState } from 'instantsearch.js'
import {
  DEFAULT_FILTER,
  getSearchFilterFromRefinement,
  type SearchFilter,
} from '@/lib/file-category'

export type SearchRouteState = {
  query?: string
  page?: number | string
  file_category?: string
}

export type SearchUiState = UiState & {
  [indexId: string]: UiState[string] & {
    query?: string
    page?: number
    refinementList?: {
      file_category?: string[]
    }
  }
}

function normalizePage(value: number | string | undefined): number | undefined {
  if (typeof value === 'number' && Number.isFinite(value) && value > 0) {
    return value
  }

  if (typeof value === 'string' && value.trim() !== '') {
    const page = Number.parseInt(value, 10)
    if (Number.isFinite(page) && page > 0) {
      return page
    }
  }

  return undefined
}

function getRouteFilter(value: string | undefined): SearchFilter {
  return getSearchFilterFromRefinement(value)
}

export function createSearchStateMapping(
  indexName: string,
): StateMapping<SearchUiState, SearchRouteState> {
  return {
    stateToRoute(uiState) {
      const indexUiState = uiState[indexName] ?? {}
      const routeState: SearchRouteState = {}
      const query = indexUiState.query?.trim()
      const page = normalizePage(indexUiState.page)
      const filter = getSearchFilterFromRefinement(indexUiState.refinementList?.file_category)

      if (query) {
        routeState.query = query
      }
      if (page && page > 1) {
        routeState.page = page
      }
      if (filter !== DEFAULT_FILTER) {
        routeState.file_category = filter
      }

      return routeState
    },
    routeToState(routeState) {
      const indexUiState: SearchUiState[string] = {}
      const query = routeState.query?.trim()
      const page = normalizePage(routeState.page)
      const filter = getRouteFilter(routeState.file_category)

      if (query) {
        indexUiState.query = query
      }
      if (page) {
        indexUiState.page = page
      }
      if (filter !== DEFAULT_FILTER) {
        indexUiState.refinementList = {
          file_category: [filter],
        }
      }

      return {
        [indexName]: indexUiState,
      }
    },
  }
}
