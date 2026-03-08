import {
  instantMeiliSearch,
  type InstantMeiliSearchObject,
} from '@meilisearch/instant-meilisearch'

export interface PublicSearchClientConfig {
  host: string
  indexName: string
  searchApiKey: string
}

export function createPublicSearchClient(
  config: PublicSearchClientConfig,
): InstantMeiliSearchObject {
  return instantMeiliSearch(config.host, config.searchApiKey)
}
