import { describe, expect, it, vi, beforeEach } from 'vitest'
import { createTypesensePublicSearchClient } from './typesense-search-client'

const fetchMock = vi.fn()

describe('createTypesensePublicSearchClient', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    vi.stubGlobal('fetch', fetchMock)
  })

  it('queries Typesense search endpoint and maps response into InstantSearch format', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({
        found: 2,
        search_time_ms: 4,
        hits: [
          {
            document: {
              doc_id: 'file_1',
              source_id: 1,
              type: 'file',
              name: 'project-design-spec.docx',
              path_text: '/project-design-spec.docx',
              parent_id: 0,
              modified_at: 1700000000,
              created_at: 1700000000,
              size: 1024,
              sha1: '',
              in_trash: false,
              is_deleted: false,
              file_category: 'doc',
            },
            highlights: [
              { field: 'name', snippet: 'project-<mark>design</mark>-spec.docx' },
            ],
          },
        ],
        facet_counts: [
          {
            field_name: 'file_category',
            counts: [{ value: 'doc', count: 2 }],
          },
        ],
      }),
    })

    const client = createTypesensePublicSearchClient({
      provider: 'typesense',
      host: 'https://typesense.example.com',
      indexName: 'npan-public',
      searchApiKey: 'public-search-key',
    })

    const result = await client.searchClient.search([{
      indexName: 'npan-public',
      params: {
        query: 'design',
        page: 0,
        hitsPerPage: 30,
        filters: '(type = "file" AND is_deleted = false AND in_trash = false) AND (file_category:"doc")',
      },
    }])

    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0] ?? []
    expect(String(url)).toContain('/collections/npan-public/documents/search')
    expect(String(url)).toContain('q=design')
    expect(String(url)).toContain('per_page=30')
    expect(String(url)).toContain('exhaustive_search=true')
    expect(String(url)).toContain('filter_by=')
    expect(String(url)).toContain('file_category%3A%3D%60doc%60')
    expect(init).toMatchObject({
      method: 'GET',
      headers: {
        'X-TYPESENSE-API-KEY': 'public-search-key',
      },
    })
    expect(result).toEqual({
      results: [
        expect.objectContaining({
          nbHits: 2,
          hitsPerPage: 30,
          page: 0,
          facets: {
            file_category: {
              doc: 2,
            },
          },
          hits: [
            expect.objectContaining({
              doc_id: 'file_1',
              _highlightResult: {
                name: { value: 'project-<mark>design</mark>-spec.docx' },
              },
            }),
          ],
        }),
      ],
    })
  })
})
