import type { Transport } from '@connectrpc/connect'
import { callUnaryMethod } from '@connectrpc/connect-query-core'
import { z } from 'zod'
import { getSearchConfig as getSearchConfigMethod } from '@/gen/npan/v1/api-AppService_connectquery'
import { appTransport } from '@/lib/connect-transport'

const PublicSearchConfigSchema = z.object({
  provider: z.enum(['meilisearch', 'typesense']).default('meilisearch'),
  host: z.string().trim(),
  indexName: z.string().trim(),
  searchApiKey: z.string().trim(),
  instantsearchEnabled: z.boolean(),
})

export type PublicSearchConfig = z.infer<typeof PublicSearchConfigSchema>
export type SearchBootstrapMode = 'public' | 'legacy'

export async function loadSearchConfig(
  transport: Transport = appTransport,
): Promise<PublicSearchConfig> {
  const response = await callUnaryMethod(transport, getSearchConfigMethod, {})

  return PublicSearchConfigSchema.parse({
    host: response.host,
    indexName: response.indexName,
    searchApiKey: response.searchApiKey,
    instantsearchEnabled: response.instantsearchEnabled,
    provider: response.provider || 'meilisearch',
  })
}

export function resolveSearchBootstrapMode(
  config: PublicSearchConfig,
): SearchBootstrapMode {
  if (!config.instantsearchEnabled) {
    return 'legacy'
  }

  if (!config.host || !config.indexName || !config.searchApiKey) {
    return 'legacy'
  }

  return 'public'
}
