import type { SearchClient } from 'algoliasearch-helper/types/algoliasearch.js'
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
  searchClient: SearchClient
}

export function createPublicSearchClient(
  config: PublicSearchClientConfig,
): PublicSearchClientInstance {
  if (config.provider === 'typesense') {
    return createTypesensePublicSearchClient(config)
  }
  return createMeiliPublicSearchClient(config)
}
