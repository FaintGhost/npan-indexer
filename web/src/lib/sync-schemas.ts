import { z } from 'zod'
import {
  zCrawlStats,
  zRootSyncProgress,
  zIncrementalSyncStats,
  zSyncProgressState,
} from '@/api/generated/zod.gen'

export const CrawlStatsSchema = zCrawlStats
export type CrawlStats = z.infer<typeof CrawlStatsSchema>

export const RootProgressSchema = zRootSyncProgress
export type RootProgress = z.infer<typeof RootProgressSchema>

export const IncrementalSyncStatsSchema = zIncrementalSyncStats
export type IncrementalSyncStats = z.infer<typeof IncrementalSyncStatsSchema>

export const SyncProgressSchema = zSyncProgressState
export type SyncProgress = z.infer<typeof SyncProgressSchema>

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
