import { z } from 'zod'

export const CrawlStatsSchema = z.object({
  foldersVisited: z.number(),
  filesIndexed: z.number(),
  pagesFetched: z.number(),
  failedRequests: z.number(),
  startedAt: z.number(),
  endedAt: z.number(),
  filesDiscovered: z.number().optional().default(0),
  skippedFiles: z.number().optional().default(0),
})

export type CrawlStats = z.infer<typeof CrawlStatsSchema>

export const RootProgressSchema = z.object({
  rootFolderId: z.number(),
  status: z.string(),
  estimatedTotalDocs: z.number().nullable().optional(),
  stats: CrawlStatsSchema,
  updatedAt: z.number(),
})

export type RootProgress = z.infer<typeof RootProgressSchema>

export const IncrementalSyncStatsSchema = z.object({
  changesFetched: z.number(),
  upserted: z.number(),
  deleted: z.number(),
  skippedUpserts: z.number(),
  skippedDeletes: z.number(),
  cursorBefore: z.number(),
  cursorAfter: z.number(),
})

export type IncrementalSyncStats = z.infer<typeof IncrementalSyncStatsSchema>

export const SyncProgressSchema = z.object({
  status: z.enum(['idle', 'running', 'done', 'error', 'cancelled', 'interrupted']),
  startedAt: z.number(),
  updatedAt: z.number(),
  roots: z.array(z.number()),
  rootNames: z.record(z.string(), z.string()).optional().default({}),
  completedRoots: z.array(z.number()),
  activeRoot: z.number().nullable().optional(),
  aggregateStats: CrawlStatsSchema,
  rootProgress: z.record(z.string(), RootProgressSchema),
  mode: z.string().optional().default(''),
  incrementalStats: IncrementalSyncStatsSchema.optional().nullable(),
  lastError: z.string().optional().default(''),
  verification: z
    .object({
      meiliDocCount: z.number(),
      crawledDocCount: z.number(),
      discoveredDocCount: z.number(),
      skippedCount: z.number(),
      verified: z.boolean(),
      warnings: z.array(z.string()).optional().default([]),
    })
    .optional()
    .nullable(),
})

export type SyncProgress = z.infer<typeof SyncProgressSchema>
