import {
  instantMeiliSearch,
} from '@meilisearch/instant-meilisearch'
import type { SearchClient } from 'algoliasearch-helper/types/algoliasearch.js'
import type { PublicSearchClientConfig, PublicSearchClientInstance } from '@/lib/public-search-client'
import { PUBLIC_BASELINE_FILTER } from '@/lib/public-search-request-adapter'

export function createMeiliPublicSearchClient(
  config: PublicSearchClientConfig,
): PublicSearchClientInstance {
  const client = instantMeiliSearch(config.host, config.searchApiKey, {
    meiliSearchParams: {
      filter: PUBLIC_BASELINE_FILTER,
    },
  })

  client.setMeiliSearchParams({
    filter: undefined,
  })

  const search: SearchClient['search'] = async <T>(
    requests: Parameters<SearchClient['search']>[0],
  ) => {
    const meiliSearch = client.searchClient.search as SearchClient['search']
    return meiliSearch<T>(requests)
  }

  return {
    searchClient: {
      search,
    },
  }
}
