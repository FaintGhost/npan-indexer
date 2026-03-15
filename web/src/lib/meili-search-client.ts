import {
  instantMeiliSearch,
  type InstantMeiliSearchObject,
} from '@meilisearch/instant-meilisearch'
import type { PublicSearchClientConfig } from '@/lib/public-search-client'
import { PUBLIC_BASELINE_FILTER } from '@/lib/public-search-request-adapter'

export function createMeiliPublicSearchClient(
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
