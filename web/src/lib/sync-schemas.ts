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
