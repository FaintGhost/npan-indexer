import { z } from 'zod'
import {
  zCrawlStats,
  zRootSyncProgress,
  zIncrementalSyncStats,
  zSyncProgressState,
} from '@/api/generated/zod.gen'

export const ProtoTimestampSchema = z.union([
  z.string(),
  z.object({
    seconds: z.union([z.number().int(), z.string(), z.bigint()]),
    nanos: z.number().int().optional(),
  }),
])
export type ProtoTimestamp = z.infer<typeof ProtoTimestampSchema>

export const CrawlStatsSchema = zCrawlStats.extend({
  startedAtTs: ProtoTimestampSchema.optional(),
  endedAtTs: ProtoTimestampSchema.optional(),
})
export type CrawlStats = z.infer<typeof CrawlStatsSchema>

export const RootProgressSchema = zRootSyncProgress.extend({
  stats: CrawlStatsSchema,
  updatedAtTs: ProtoTimestampSchema.optional(),
})
export type RootProgress = z.infer<typeof RootProgressSchema>

export const IncrementalSyncStatsSchema = zIncrementalSyncStats
export type IncrementalSyncStats = z.infer<typeof IncrementalSyncStatsSchema>

export const SyncProgressSchema = zSyncProgressState.extend({
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
