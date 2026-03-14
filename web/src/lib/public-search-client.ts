import type { InstantMeiliSearchObject } from '@meilisearch/instant-meilisearch'
import { createMeiliPublicSearchClient } from '@/lib/meili-search-client'
import { createTypesensePublicSearchClient } from '@/lib/typesense-search-client'

export type PublicSearchProvider = 'meilisearch' | 'typesense'

export interface PublicSearchClientConfig {
  provider: PublicSearchProvider
  host: string
  indexName: string
  searchApiKey: string
}

export interface PublicSearchClientInstance {
  searchClient: {
    search: (requests: Array<Record<string, unknown>>) => Promise<{ results: Array<Record<string, unknown>> }>
  }
}

export function createPublicSearchClient(
  config: PublicSearchClientConfig,
): InstantMeiliSearchObject | PublicSearchClientInstance {
  if (config.provider === 'typesense') {
    return createTypesensePublicSearchClient(config)
  }
  return createMeiliPublicSearchClient(config)
}
