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

  it('parses data with filesDiscovered and skippedFiles', () => {
    const result = CrawlStatsSchema.safeParse({
      ...validStats,
      filesDiscovered: 500,
      skippedFiles: 10,
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.filesDiscovered).toBe(500)
      expect(result.data.skippedFiles).toBe(10)
    }
  })

  it('defaults filesDiscovered and skippedFiles to 0 when missing', () => {
    const result = CrawlStatsSchema.safeParse(validStats)
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.filesDiscovered).toBe(0)
      expect(result.data.skippedFiles).toBe(0)
    }
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

  it('parses data with verification field', () => {
    const result = SyncProgressSchema.safeParse({
      ...validProgress,
      verification: {
        meiliDocCount: 1000,
        crawledDocCount: 950,
        discoveredDocCount: 1100,
        skippedCount: 50,
        verified: true,
        warnings: ['some warning'],
      },
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.verification).not.toBeNull()
      expect(result.data.verification?.meiliDocCount).toBe(1000)
      expect(result.data.verification?.crawledDocCount).toBe(950)
      expect(result.data.verification?.discoveredDocCount).toBe(1100)
      expect(result.data.verification?.skippedCount).toBe(50)
      expect(result.data.verification?.verified).toBe(true)
      expect(result.data.verification?.warnings).toEqual(['some warning'])
    }
  })

  it('parses data without verification field (backward compat)', () => {
    const result = SyncProgressSchema.safeParse(validProgress)
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.verification).toBeUndefined()
    }
  })

  it('parses data with null verification field', () => {
    const result = SyncProgressSchema.safeParse({
      ...validProgress,
      verification: null,
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.verification).toBeNull()
    }
  })

  it('defaults verification warnings to empty array when missing', () => {
    const result = SyncProgressSchema.safeParse({
      ...validProgress,
      verification: {
        meiliDocCount: 100,
        crawledDocCount: 100,
        discoveredDocCount: 100,
        skippedCount: 0,
        verified: true,
      },
    })
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.verification?.warnings).toEqual([])
    }
  })
})
