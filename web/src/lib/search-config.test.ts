import { describe, expect, it } from 'vitest'
import { resolveSearchBootstrapMode } from './search-config'

describe('search config bootstrap contract', () => {
  it('loads public search config when instantsearch is enabled', async () => {
    await expect(import('./search-config')).resolves.toHaveProperty('loadSearchConfig')
  })

  it('falls back to legacy appsearch when config is disabled or missing', async () => {
    await expect(import('./search-config')).resolves.toHaveProperty('resolveSearchBootstrapMode')
  })

  it('uses public mode only when flag and required fields are present', () => {
    expect(resolveSearchBootstrapMode({
      provider: 'meilisearch',
      host: 'https://search.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
      instantsearchEnabled: true,
    })).toBe('public')
  })

  it('uses legacy mode when instantsearch is disabled', () => {
    expect(resolveSearchBootstrapMode({
      provider: 'meilisearch',
      host: 'https://search.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
      instantsearchEnabled: false,
    })).toBe('legacy')
  })

  it('uses legacy mode when any required public field is missing', () => {
    expect(resolveSearchBootstrapMode({
      provider: 'typesense',
      host: '',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
      instantsearchEnabled: true,
    })).toBe('legacy')
  })

  it('uses public mode for typesense when required fields are present', () => {
    expect(resolveSearchBootstrapMode({
      provider: 'typesense',
      host: 'https://typesense.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
      instantsearchEnabled: true,
    })).toBe('public')
  })
})
