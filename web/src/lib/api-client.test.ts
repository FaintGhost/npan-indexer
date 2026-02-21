import { describe, it, expect } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { apiGet, apiPost, ApiError } from './api-client'
import { SearchResponseSchema } from './schemas'

describe('apiGet', () => {
  it('fetches and validates response with schema', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json({
          items: [{
            doc_id: 'file_1',
            source_id: 1,
            type: 'file',
            name: 'test.txt',
            path_text: '/test.txt',
            parent_id: 0,
            modified_at: 1700000000,
            created_at: 1700000000,
            size: 100,
            sha1: 'abc',
            in_trash: false,
            is_deleted: false,
          }],
          total: 1,
        })
      }),
    )

    const result = await apiGet('/api/v1/app/search', { query: 'test' }, SearchResponseSchema)
    expect(result.total).toBe(1)
    expect(result.items).toHaveLength(1)
    expect(result.items[0]!.name).toBe('test.txt')
  })

  it('throws ApiError on HTTP error', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json(
          { code: 'BAD_REQUEST', message: '缺少 query 参数' },
          { status: 400 },
        )
      }),
    )

    await expect(
      apiGet('/api/v1/app/search', {}, SearchResponseSchema),
    ).rejects.toThrow(ApiError)

    try {
      await apiGet('/api/v1/app/search', {}, SearchResponseSchema)
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError)
      const err = e as ApiError
      expect(err.status).toBe(400)
      expect(err.code).toBe('BAD_REQUEST')
      expect(err.message).toBe('缺少 query 参数')
    }
  })

  it('throws on schema validation failure', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json({ unexpected: 'data' })
      }),
    )

    await expect(
      apiGet('/api/v1/app/search', {}, SearchResponseSchema),
    ).rejects.toThrow()
  })

  it('supports abort signal', async () => {
    server.use(
      http.get('/api/v1/app/search', async () => {
        await new Promise((r) => setTimeout(r, 5000))
        return HttpResponse.json({ items: [], total: 0 })
      }),
    )

    const controller = new AbortController()
    controller.abort()

    await expect(
      apiGet('/api/v1/app/search', { query: 'test' }, SearchResponseSchema, {
        signal: controller.signal,
      }),
    ).rejects.toThrow()
  })

  it('filters out undefined/null/empty params', async () => {
    let capturedUrl = ''
    server.use(
      http.get('/api/v1/app/search', ({ request }) => {
        capturedUrl = request.url
        return HttpResponse.json({ items: [], total: 0 })
      }),
    )

    await apiGet(
      '/api/v1/app/search',
      { query: 'test', page: undefined, empty: '', valid: 1 },
      SearchResponseSchema,
    )

    const url = new URL(capturedUrl)
    expect(url.searchParams.get('query')).toBe('test')
    expect(url.searchParams.get('valid')).toBe('1')
    expect(url.searchParams.has('page')).toBe(false)
    expect(url.searchParams.has('empty')).toBe(false)
  })
})

describe('apiPost', () => {
  it('sends POST with JSON body and custom headers', async () => {
    let capturedHeaders: Record<string, string> = {}
    let capturedBody: unknown = null

    server.use(
      http.post('/api/v1/admin/sync', async ({ request }) => {
        capturedHeaders = Object.fromEntries(request.headers.entries())
        capturedBody = await request.json()
        return HttpResponse.json({ message: 'Sync started' })
      }),
    )

    const result = await apiPost(
      '/api/v1/admin/sync',
      { root_folder_ids: [100, 200] },
      { headers: { 'X-API-Key': 'test-key' } },
    )

    expect(result.message).toBe('Sync started')
    expect(capturedHeaders['x-api-key']).toBe('test-key')
    expect(capturedBody).toEqual({ root_folder_ids: [100, 200] })
  })

  it('throws ApiError on HTTP error', async () => {
    server.use(
      http.post('/api/v1/admin/sync', () => {
        return HttpResponse.json(
          { code: 'UNAUTHORIZED', message: 'Invalid API key' },
          { status: 401 },
        )
      }),
    )

    await expect(
      apiPost('/api/v1/admin/sync', {}, { headers: { 'X-API-Key': 'bad' } }),
    ).rejects.toThrow(ApiError)
  })
})
