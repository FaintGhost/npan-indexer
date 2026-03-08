import {
  instantMeiliSearch,
  type InstantMeiliSearchObject,
} from '@meilisearch/instant-meilisearch'
import { PUBLIC_BASELINE_FILTER } from '@/lib/public-search-request-adapter'

export interface PublicSearchClientConfig {
  host: string
  indexName: string
  searchApiKey: string
}

export function createPublicSearchClient(
  config: PublicSearchClientConfig,
): InstantMeiliSearchObject {
  const client = instantMeiliSearch(config.host, config.searchApiKey, {
    meiliSearchParams: {
      filter: PUBLIC_BASELINE_FILTER,
    },
  })

  client.setMeiliSearchParams({
    filter: undefined,
  })

  return client
}
