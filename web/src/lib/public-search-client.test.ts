import { beforeEach, describe, expect, it, vi } from 'vitest'

const { createMeiliPublicSearchClientMock, createTypesensePublicSearchClientMock } = vi.hoisted(() => ({
  createMeiliPublicSearchClientMock: vi.fn(),
  createTypesensePublicSearchClientMock: vi.fn(),
}))

vi.mock('./meili-search-client', () => ({
  createMeiliPublicSearchClient: createMeiliPublicSearchClientMock,
}))

vi.mock('./typesense-search-client', () => ({
  createTypesensePublicSearchClient: createTypesensePublicSearchClientMock,
}))

import { createPublicSearchClient } from './public-search-client'

describe('createPublicSearchClient', () => {
  beforeEach(() => {
    createMeiliPublicSearchClientMock.mockReset()
    createTypesensePublicSearchClientMock.mockReset()
    createMeiliPublicSearchClientMock.mockReturnValue({ searchClient: {} })
    createTypesensePublicSearchClientMock.mockReturnValue({ searchClient: {} })
  })

  it('dispatches meilisearch configs to the Meilisearch client factory', () => {
    createPublicSearchClient({
      provider: 'meilisearch',
      host: 'https://search.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
    })

    expect(createMeiliPublicSearchClientMock).toHaveBeenCalledWith({
      provider: 'meilisearch',
      host: 'https://search.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
    })
    expect(createTypesensePublicSearchClientMock).not.toHaveBeenCalled()
  })

  it('dispatches typesense configs to the Typesense client factory', () => {
    createPublicSearchClient({
      provider: 'typesense',
      host: 'https://typesense.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
    })

    expect(createTypesensePublicSearchClientMock).toHaveBeenCalledWith({
      provider: 'typesense',
      host: 'https://typesense.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
    })
    expect(createMeiliPublicSearchClientMock).not.toHaveBeenCalled()
  })
})
