import type { PublicSearchClientConfig, PublicSearchClientInstance } from '@/lib/public-search-client'

type PublicSearchRequest = {
  indexName?: unknown
  params?: Record<string, unknown>
  query?: unknown
}

type TypesenseFacetCount = {
  field_name?: string
  counts?: Array<{ value?: string; count?: number }>
}

type TypesenseSearchResponse = {
  found?: number
  search_time_ms?: number
  hits?: Array<{
    document?: Record<string, unknown>
    highlights?: Array<{ field?: string; snippet?: string; value?: string }>
  }>
  facet_counts?: TypesenseFacetCount[]
}

function readString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function readPositiveInt(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0
    ? Math.floor(value)
    : fallback
}

function getQuery(request: PublicSearchRequest): string {
  if (typeof request.params?.query === 'string') {
    return request.params.query
  }
  if (typeof request.query === 'string') {
    return request.query
  }
  return ''
}

function getFilters(request: PublicSearchRequest): string {
  const params = request.params ?? {}
  if (typeof params.filters === 'string') {
    return params.filters
  }
  if (typeof params.filter === 'string') {
    return params.filter
  }
  return ''
}

function normalizeTypesenseFilter(filter: string): string {
  return filter
    .replace(/\bAND\b/g, '&&')
    .replace(/\bOR\b/g, '||')
    .replace(/(\w+)\s*=\s*"([^"]+)"/g, '$1:=`$2`')
    .replace(/(\w+)\s*=\s*(true|false|\d+)/g, '$1:=$2')
    .replace(/(\w+):"([^"]+)"/g, '$1:=`$2`')
    .replace(/(\w+):([A-Za-z0-9_-]+)/g, '$1:=`$2`')
}

function toFacetMap(facetCounts: TypesenseFacetCount[] | undefined): Record<string, Record<string, number>> {
  const result: Record<string, Record<string, number>> = {}
  for (const facet of facetCounts ?? []) {
    const fieldName = readString(facet.field_name)
    if (!fieldName) {
      continue
    }
    result[fieldName] = {}
    for (const count of facet.counts ?? []) {
      const value = readString(count.value)
      if (!value) {
        continue
      }
      result[fieldName][value] = typeof count.count === 'number' ? count.count : 0
    }
  }
  return result
}

function toHighlightResult(hit: NonNullable<TypesenseSearchResponse['hits']>[number]): Record<string, { value: string }> | undefined {
  for (const highlight of hit.highlights ?? []) {
    if (highlight.field === 'name') {
      const value = readString(highlight.snippet) || readString(highlight.value)
      if (value) {
        return {
          name: { value },
        }
      }
    }
  }
  return undefined
}

function buildTypesenseURL(config: PublicSearchClientConfig, request: PublicSearchRequest): string {
  const params = request.params ?? {}
  const url = new URL(`/collections/${config.indexName}/documents/search`, config.host)
  const page = readPositiveInt(params.page, 0) + 1
  const perPage = readPositiveInt(params.hitsPerPage, 20)

  url.searchParams.set('q', getQuery(request) || '*')
  url.searchParams.set('query_by', 'name_base,name_ext,name,path_text')
  url.searchParams.set('query_by_weights', '8,6,4,1')
  url.searchParams.set('highlight_fields', 'name')
  url.searchParams.set('highlight_full_fields', 'name')
  url.searchParams.set('highlight_start_tag', '<mark>')
  url.searchParams.set('highlight_end_tag', '</mark>')
  url.searchParams.set('facet_by', 'file_category')
  url.searchParams.set('sort_by', '_text_match:desc,modified_at:desc')
  url.searchParams.set('per_page', String(perPage))
  url.searchParams.set('page', String(page))
  url.searchParams.set('exhaustive_search', 'true')

  const filter = normalizeTypesenseFilter(getFilters(request))
  if (filter.trim()) {
    url.searchParams.set('filter_by', filter)
  }

  return url.toString()
}

export function createTypesensePublicSearchClient(
  config: PublicSearchClientConfig,
): PublicSearchClientInstance {
  return {
    searchClient: {
      async search(requests: PublicSearchRequest[]) {
        const results = await Promise.all(requests.map(async (request) => {
          const response = await fetch(buildTypesenseURL(config, request), {
            method: 'GET',
            headers: {
              'X-TYPESENSE-API-KEY': config.searchApiKey,
            },
          })
          if (!response.ok) {
            throw new Error(`Typesense public search failed: ${response.status}`)
          }

          const payload = await response.json() as TypesenseSearchResponse
          const hitsPerPage = readPositiveInt(request.params?.hitsPerPage, 20)
          const page = readPositiveInt(request.params?.page, 0)
          const nbHits = typeof payload.found === 'number' ? payload.found : 0

          return {
            index: readString(request.indexName) || config.indexName,
            hitsPerPage,
            page,
            nbPages: Math.max(1, Math.ceil(nbHits / hitsPerPage)),
            nbHits,
            processingTimeMS: typeof payload.search_time_ms === 'number' ? payload.search_time_ms : 0,
            query: getQuery(request).trim(),
            hits: (payload.hits ?? []).map((hit) => ({
              ...(hit.document ?? {}),
              _highlightResult: toHighlightResult(hit),
            })),
            exhaustiveNbHits: true,
            facets: toFacetMap(payload.facet_counts),
            facets_stats: {},
            params: '',
          }
        }))

        return { results }
      },
    },
  }
}
