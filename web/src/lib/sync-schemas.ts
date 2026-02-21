import { z } from 'zod'

export const CrawlStatsSchema = z.object({
  foldersVisited: z.number(),
  filesIndexed: z.number(),
  pagesFetched: z.number(),
  failedRequests: z.number(),
  startedAt: z.number(),
  endedAt: z.number(),
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

export const SyncProgressSchema = z.object({
  status: z.enum(['idle', 'running', 'done', 'error', 'cancelled']),
  startedAt: z.number(),
  updatedAt: z.number(),
  roots: z.array(z.number()),
  rootNames: z.record(z.string(), z.string()).optional().default({}),
  completedRoots: z.array(z.number()),
  activeRoot: z.number().nullable().optional(),
  aggregateStats: CrawlStatsSchema,
  rootProgress: z.record(z.string(), RootProgressSchema),
  lastError: z.string().optional().default(''),
})

export type SyncProgress = z.infer<typeof SyncProgressSchema>
