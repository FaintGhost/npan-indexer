import type { InstantMeiliSearchObject } from '@meilisearch/instant-meilisearch'
import { normalizeSearchQuery } from '@/lib/search-query-normalizer'

export const PUBLIC_BASELINE_FILTER = 'type = "file" AND is_deleted = false AND in_trash = false'

type PublicSearchParams = Record<string, unknown> & {
  query?: unknown
  filter?: unknown
  filters?: unknown
}

type PublicSearchRequest = {
  query?: unknown
  params?: PublicSearchParams
  [key: string]: unknown
}

function getRawQuery(request: PublicSearchRequest): string | undefined {
  if (typeof request.query === 'string' && request.query.trim() !== '') {
    return request.query
  }

  if (typeof request.params?.query === 'string' && request.params.query.trim() !== '') {
    return request.params.query
  }

  if (typeof request.query === 'string') {
    return request.query
  }

  if (typeof request.params?.query === 'string') {
    return request.params.query
  }

  return undefined
}

function getExistingFilter(params?: PublicSearchParams): string | undefined {
  if (typeof params?.filters === 'string' && params.filters.trim()) {
    return params.filters.trim()
  }

  if (typeof params?.filter === 'string' && params.filter.trim()) {
    return params.filter.trim()
  }

  return undefined
}

function mergePublicBaselineFilter(existingFilter?: string): string {
  if (!existingFilter) {
    return PUBLIC_BASELINE_FILTER
  }

  return `(${PUBLIC_BASELINE_FILTER}) AND (${existingFilter})`
}

function hasNonEmptyQuery(request: PublicSearchRequest): boolean {
  const rawQuery = getRawQuery(request)
  return typeof rawQuery === 'string' && rawQuery.trim() !== ''
}

function normalizePositiveNumber(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0
    ? value
    : fallback
}

function normalizeNonNegativeNumber(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0
    ? value
    : fallback
}

function createEmptySearchResponse(requests: PublicSearchRequest[]) {
  return {
    results: requests.map((request) => {
      const rawQuery = getRawQuery(request)

      return {
        index: typeof request.indexName === 'string' ? request.indexName : '',
        hitsPerPage: normalizePositiveNumber(request.params?.hitsPerPage, 20),
        page: normalizeNonNegativeNumber(request.params?.page, 0),
        facets: {},
        nbPages: 1,
        nbHits: 0,
        processingTimeMS: 0,
        query: typeof rawQuery === 'string' ? rawQuery.trim() : '',
        hits: [],
        params: '',
        exhaustiveNbHits: true,
        facets_stats: {},
      }
    }),
  }
}

export function adaptPublicSearchRequest<T extends PublicSearchRequest>(request: T): T {
  const rawQuery = getRawQuery(request)
  const normalizedQuery = rawQuery === undefined
    ? undefined
    : normalizeSearchQuery(rawQuery)

  return {
    ...request,
    query: normalizedQuery ?? request.query,
    params: {
      ...(request.params ?? {}),
      ...(normalizedQuery === undefined ? {} : { query: normalizedQuery }),
      filters: mergePublicBaselineFilter(getExistingFilter(request.params)),
    },
  }
}

export function wrapPublicSearchClient<T extends InstantMeiliSearchObject>(client: T): T {
  const search = client.searchClient.search.bind(client.searchClient)

  return {
    ...client,
    searchClient: {
      ...client.searchClient,
      search(requests) {
        if (!requests.some((request) => hasNonEmptyQuery(request))) {
          return Promise.resolve(createEmptySearchResponse(requests))
        }

        return search(requests.map((request) => adaptPublicSearchRequest(request)))
      },
    },
  }
}
