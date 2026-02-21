import { describe, it, expect } from 'vitest'
import {
  CrawlStatsSchema,
  RootProgressSchema,
  SyncProgressSchema,
} from './sync-schemas'

const validStats = {
  foldersVisited: 10,
  filesIndexed: 100,
  pagesFetched: 20,
  failedRequests: 0,
  startedAt: 1700000000,
  endedAt: 1700001000,
}

describe('CrawlStatsSchema', () => {
  it('parses valid stats', () => {
    const result = CrawlStatsSchema.safeParse(validStats)
    expect(result.success).toBe(true)
  })

  it('rejects non-number fields', () => {
    const result = CrawlStatsSchema.safeParse({
      ...validStats,
      filesIndexed: 'not a number',
    })
    expect(result.success).toBe(false)
  })
})

describe('RootProgressSchema', () => {
  it('parses valid root progress', () => {
    const result = RootProgressSchema.safeParse({
      rootFolderId: 12345,
      status: 'running',
      stats: validStats,
      updatedAt: 1700000500,
    })
    expect(result.success).toBe(true)
  })

  it('parses with optional estimatedTotalDocs', () => {
    const result = RootProgressSchema.safeParse({
      rootFolderId: 12345,
      status: 'done',
      estimatedTotalDocs: 500,
      stats: validStats,
      updatedAt: 1700000500,
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.estimatedTotalDocs).toBe(500)
    }
  })

  it('parses with null estimatedTotalDocs', () => {
    const result = RootProgressSchema.safeParse({
      rootFolderId: 12345,
      status: 'running',
      estimatedTotalDocs: null,
      stats: validStats,
      updatedAt: 1700000500,
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.estimatedTotalDocs).toBeNull()
    }
  })
})

describe('SyncProgressSchema', () => {
  const validProgress = {
    status: 'running',
    startedAt: 1700000000,
    updatedAt: 1700000500,
    roots: [100, 200],
    completedRoots: [100],
    aggregateStats: validStats,
    rootProgress: {
      '100': {
        rootFolderId: 100,
        status: 'done',
        stats: validStats,
        updatedAt: 1700000500,
      },
    },
  }

  it('parses valid sync progress', () => {
    const result = SyncProgressSchema.safeParse(validProgress)
    expect(result.success).toBe(true)
  })

  it('accepts valid status values', () => {
    for (const status of ['idle', 'running', 'done', 'error', 'cancelled']) {
      const result = SyncProgressSchema.safeParse({ ...validProgress, status })
      expect(result.success).toBe(true)
    }
  })

  it('parses with optional activeRoot', () => {
    const result = SyncProgressSchema.safeParse({
      ...validProgress,
      activeRoot: 200,
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.activeRoot).toBe(200)
    }
  })

  it('parses with null activeRoot', () => {
    const result = SyncProgressSchema.safeParse({
      ...validProgress,
      activeRoot: null,
    })
    expect(result.success).toBe(true)
  })

  it('defaults lastError to empty string when missing', () => {
    const result = SyncProgressSchema.safeParse(validProgress)
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.lastError).toBe('')
    }
  })

  it('parses lastError when present', () => {
    const result = SyncProgressSchema.safeParse({
      ...validProgress,
      lastError: 'network timeout',
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.lastError).toBe('network timeout')
    }
  })

  it('parses empty rootProgress', () => {
    const result = SyncProgressSchema.safeParse({
      ...validProgress,
      rootProgress: {},
    })
    expect(result.success).toBe(true)
  })
})
