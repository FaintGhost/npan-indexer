import { beforeEach, describe, expect, it, vi } from 'vitest'

const { instantMeiliSearchMock } = vi.hoisted(() => ({
  instantMeiliSearchMock: vi.fn(),
}))

vi.mock('@meilisearch/instant-meilisearch', () => ({
  instantMeiliSearch: instantMeiliSearchMock,
}))

import { createPublicSearchClient } from './meili-search-client'

describe('createPublicSearchClient', () => {
  beforeEach(() => {
    instantMeiliSearchMock.mockReset()
    instantMeiliSearchMock.mockReturnValue({
      searchClient: {},
      setMeiliSearchParams: vi.fn(),
      meiliSearchInstance: {},
    })
  })

  it('passes public baseline filters through the official instant-meilisearch configuration entry', () => {
    createPublicSearchClient({
      host: 'https://search.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
    })

    expect(instantMeiliSearchMock).toHaveBeenCalledTimes(1)

    const options = instantMeiliSearchMock.mock.calls[0]?.[2]
    expect(options, '当前 public 搜索 client 未通过官方配置入口注入默认过滤基线').toBeDefined()
    expect(options).toEqual(
      expect.objectContaining({
        meiliSearchParams: expect.objectContaining({
          filter: expect.stringMatching(/type\s*(?:=|:)\s*["']?file["']?/i),
        }),
      }),
    )

    const filter = options?.meiliSearchParams?.filter as string | undefined
    expect(filter, '当前 public 搜索 client 缺少默认过滤基线').toBeDefined()
    expect(filter).toMatch(/is_deleted\s*(?:=|:)\s*(?:false|0)/i)
    expect(filter).toMatch(/in_trash\s*(?:=|:)\s*(?:false|0)/i)
  })

  it('keeps default filters in a dedicated baseline config so file_category refinement can only compose on top', () => {
    createPublicSearchClient({
      host: 'https://search.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
    })

    const options = instantMeiliSearchMock.mock.calls[0]?.[2]
    const filter = options?.meiliSearchParams?.filter as string | undefined

    expect(filter, '当前 public 搜索 client 缺少默认过滤基线').toBeDefined()
    expect(filter).not.toMatch(/file_category\s*(?:=|:)/i)
  })
})
