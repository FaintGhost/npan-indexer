import { z } from 'zod'

export const ProtoTimestampSchema = z.union([
  z.string(),
  z.object({
    seconds: z.union([z.number().int(), z.string(), z.bigint()]),
    nanos: z.number().int().optional(),
  }),
])
export type ProtoTimestamp = z.infer<typeof ProtoTimestampSchema>

export const CrawlStatsSchema = z.object({
  foldersVisited: z.number().int(),
  filesIndexed: z.number().int(),
  filesDiscovered: z.number().int(),
  skippedFiles: z.number().int(),
  pagesFetched: z.number().int(),
  failedRequests: z.number().int(),
  startedAt: z.number().int(),
  endedAt: z.number().int(),
}).extend({
  startedAtTs: ProtoTimestampSchema.optional(),
  endedAtTs: ProtoTimestampSchema.optional(),
})
export type CrawlStats = z.infer<typeof CrawlStatsSchema>

export const RootProgressSchema = z.object({
  rootFolderId: z.number().int(),
  status: z.string(),
  itemCount: z.number().int().nullable().optional(),
  estimatedTotalDocs: z.number().int().nullable().optional(),
  currentFolderId: z.number().int().nullable().optional(),
  currentPageId: z.number().int().nullable().optional(),
  currentPageCount: z.number().int().nullable().optional(),
  queueLength: z.number().int().nullable().optional(),
  updatedAt: z.number().int(),
  error: z.string().optional(),
}).extend({
  stats: CrawlStatsSchema,
  updatedAtTs: ProtoTimestampSchema.optional(),
})
export type RootProgress = z.infer<typeof RootProgressSchema>

export const IncrementalSyncStatsSchema = z.object({
  changesFetched: z.number().int(),
  upserted: z.number().int(),
  deleted: z.number().int(),
  skippedUpserts: z.number().int(),
  skippedDeletes: z.number().int(),
  cursorBefore: z.number().int(),
  cursorAfter: z.number().int(),
})
export type IncrementalSyncStats = z.infer<typeof IncrementalSyncStatsSchema>

const SyncVerificationSchema = z.object({
  meiliDocCount: z.number().int(),
  crawledDocCount: z.number().int(),
  discoveredDocCount: z.number().int(),
  skippedCount: z.number().int(),
  verified: z.boolean(),
  warnings: z.array(z.string()).optional(),
})

export const SyncProgressSchema = z.object({
  status: z.enum(['idle', 'running', 'done', 'error', 'cancelled', 'interrupted']),
  mode: z.enum(['full', 'incremental']).optional(),
  startedAt: z.number().int(),
  updatedAt: z.number().int(),
  roots: z.array(z.number().int()),
  rootNames: z.record(z.string()).optional(),
  catalogRoots: z.array(z.number().int()).optional(),
  catalogRootNames: z.record(z.string()).optional(),
  completedRoots: z.array(z.number().int()),
  activeRoot: z.number().int().nullable().optional(),
  lastError: z.string().optional(),
  verification: SyncVerificationSchema.nullable().optional(),
  incrementalStats: IncrementalSyncStatsSchema.optional(),
}).extend({
  startedAtTs: ProtoTimestampSchema.optional(),
  updatedAtTs: ProtoTimestampSchema.optional(),
  aggregateStats: CrawlStatsSchema,
  rootProgress: z.record(RootProgressSchema),
  catalogRootProgress: z.record(RootProgressSchema).optional(),
})
export type SyncProgress = z.infer<typeof SyncProgressSchema>

function secondsToMillis(seconds: number | string | bigint, nanos: number): number | null {
  if (typeof seconds === 'bigint') {
    return Number(seconds * 1000n) + Math.trunc(nanos / 1_000_000)
  }
  const parsedSeconds = typeof seconds === 'string' ? Number(seconds) : seconds
  if (!Number.isFinite(parsedSeconds)) {
    return null
  }
  return Math.trunc(parsedSeconds * 1000) + Math.trunc(nanos / 1_000_000)
}

export function timestampLikeToMillis(value: unknown): number | null {
  if (value == null) {
    return null
  }

  if (typeof value === 'string') {
    const parsed = Date.parse(value)
    return Number.isFinite(parsed) ? parsed : null
  }

  if (typeof value !== 'object') {
    return null
  }

  const record = value as Record<string, unknown>
  if (!('seconds' in record)) {
    return null
  }
  const nanosRaw = typeof record.nanos === 'number' ? record.nanos : 0
  const seconds = record.seconds
  if (
    typeof seconds !== 'number' &&
    typeof seconds !== 'string' &&
    typeof seconds !== 'bigint'
  ) {
    return null
  }
  return secondsToMillis(seconds, nanosRaw)
}

export function preferTimestampMillis(legacyMillis: number, timestamp: unknown): number {
  return timestampLikeToMillis(timestamp) ?? legacyMillis
}

export const InspectRootItemSchema = z.object({
  folder_id: z.number().int(),
  name: z.string(),
  item_count: z.number().int(),
  estimated_total_docs: z.number().int(),
})
export type InspectRootItem = z.infer<typeof InspectRootItemSchema>

export const InspectRootErrorSchema = z.object({
  folder_id: z.number().int(),
  message: z.string(),
})
export type InspectRootError = z.infer<typeof InspectRootErrorSchema>

export const InspectRootsResponseSchema = z.object({
  items: z.array(InspectRootItemSchema),
  errors: z.array(InspectRootErrorSchema).optional().default([]),
})
export type InspectRootsResponse = z.infer<typeof InspectRootsResponseSchema>
