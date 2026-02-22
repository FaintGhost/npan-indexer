import { describe, it, expect } from 'vitest'
import {
  SearchResponseSchema,
  DownloadURLResponseSchema,
  ErrorResponseSchema,
  IndexDocumentSchema,
} from './schemas'

describe('IndexDocumentSchema', () => {
  const validDoc = {
    doc_id: 'file_123',
    source_id: 456,
    type: 'file',
    name: 'report.pdf',
    path_text: '/docs/report.pdf',
    parent_id: 10,
    modified_at: 1700000000,
    created_at: 1700000000,
    size: 1024,
    sha1: 'abc123',
    in_trash: false,
    is_deleted: false,
  }

  it('parses valid document', () => {
    const result = IndexDocumentSchema.safeParse(validDoc)
    expect(result.success).toBe(true)
  })

  it('parses with optional highlighted_name', () => {
    const result = IndexDocumentSchema.safeParse({
      ...validDoc,
      highlighted_name: '<mark>report</mark>.pdf',
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.highlighted_name).toBe('<mark>report</mark>.pdf')
    }
  })

  it('highlighted_name is undefined when missing', () => {
    const result = IndexDocumentSchema.safeParse(validDoc)
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.highlighted_name).toBeUndefined()
    }
  })

  it('rejects missing required fields', () => {
    const result = IndexDocumentSchema.safeParse({ doc_id: 'x' })
    expect(result.success).toBe(false)
  })
})

describe('SearchResponseSchema', () => {
  it('parses valid response with items', () => {
    const result = SearchResponseSchema.safeParse({
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
    expect(result.success).toBe(true)
  })

  it('parses empty items array', () => {
    const result = SearchResponseSchema.safeParse({ items: [], total: 0 })
    expect(result.success).toBe(true)
  })

  it('rejects missing total', () => {
    const result = SearchResponseSchema.safeParse({ items: [] })
    expect(result.success).toBe(false)
  })
})

describe('DownloadURLResponseSchema', () => {
  it('parses valid response', () => {
    const result = DownloadURLResponseSchema.safeParse({
      file_id: 123,
      download_url: 'https://example.com/file.pdf',
    })
    expect(result.success).toBe(true)
  })

  it('rejects empty download_url', () => {
    const result = DownloadURLResponseSchema.safeParse({
      file_id: 123,
      download_url: '',
    })
    expect(result.success).toBe(false)
  })

  it('rejects missing file_id', () => {
    const result = DownloadURLResponseSchema.safeParse({
      download_url: 'https://example.com/file.pdf',
    })
    expect(result.success).toBe(false)
  })
})

describe('ErrorResponseSchema', () => {
  it('parses valid error response', () => {
    const result = ErrorResponseSchema.safeParse({
      code: 'BAD_REQUEST',
      message: 'Missing query',
    })
    expect(result.success).toBe(true)
  })

  it('parses with optional request_id', () => {
    const result = ErrorResponseSchema.safeParse({
      code: 'INTERNAL_ERROR',
      message: 'Server error',
      request_id: 'req-123',
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.request_id).toBe('req-123')
    }
  })
})
