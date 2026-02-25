import type { Timestamp } from '@bufbuild/protobuf/wkt'
import type {
  CrawlStats as ProtoCrawlStats,
  GetSyncProgressResponse,
  InspectRootsResponse as ProtoInspectRootsResponse,
  RootSyncProgress as ProtoRootSyncProgress,
  SyncProgressState as ProtoSyncProgressState,
} from '@/gen/npan/v1/api_pb'
import { SyncMode, SyncStatus } from '@/gen/npan/v1/api_pb'
import type {
  CrawlStats,
  InspectRootsResponse,
  RootProgress,
  SyncProgress,
} from '@/lib/sync-schemas'

function int64ToNumber(value: bigint | undefined): number {
  if (value == null) {
    return 0
  }
  return Number(value)
}

function timestampToProtoLike(value: Timestamp | undefined) {
  if (!value) {
    return undefined
  }
  return {
    seconds: value.seconds,
    nanos: value.nanos,
  }
}

function mapSyncStatus(status: SyncStatus): SyncProgress['status'] {
  switch (status) {
    case SyncStatus.RUNNING:
      return 'running'
    case SyncStatus.DONE:
      return 'done'
    case SyncStatus.ERROR:
      return 'error'
    case SyncStatus.CANCELLED:
      return 'cancelled'
    case SyncStatus.INTERRUPTED:
      return 'interrupted'
    case SyncStatus.IDLE:
    case SyncStatus.UNSPECIFIED:
    default:
      return 'idle'
  }
}

function mapSyncMode(mode: SyncMode | undefined): SyncProgress['mode'] {
  switch (mode) {
    case SyncMode.FULL:
      return 'full'
    case SyncMode.INCREMENTAL:
      return 'incremental'
    case SyncMode.AUTO:
      return 'auto'
    default:
      return undefined
  }
}

function mapCrawlStats(stats: ProtoCrawlStats | undefined): CrawlStats {
  return {
    foldersVisited: int64ToNumber(stats?.foldersVisited),
    filesIndexed: int64ToNumber(stats?.filesIndexed),
    filesDiscovered: int64ToNumber(stats?.filesDiscovered),
    skippedFiles: int64ToNumber(stats?.skippedFiles),
    pagesFetched: int64ToNumber(stats?.pagesFetched),
    failedRequests: int64ToNumber(stats?.failedRequests),
    startedAt: int64ToNumber(stats?.startedAt),
    endedAt: int64ToNumber(stats?.endedAt),
    startedAtTs: timestampToProtoLike(stats?.startedAtTs),
    endedAtTs: timestampToProtoLike(stats?.endedAtTs),
  }
}

function mapRootProgress(value: ProtoRootSyncProgress | undefined): RootProgress {
  return {
    rootFolderId: int64ToNumber(value?.rootFolderId),
    status: value?.status ?? 'pending',
    estimatedTotalDocs:
      value?.estimatedTotalDocs != null
        ? int64ToNumber(value.estimatedTotalDocs)
        : null,
    stats: mapCrawlStats(value?.stats),
    updatedAt: int64ToNumber(value?.updatedAt),
    updatedAtTs: timestampToProtoLike(value?.updatedAtTs),
  }
}

function mapRootProgressMap(
  source: Record<string, ProtoRootSyncProgress>,
): Record<string, RootProgress> {
  const result: Record<string, RootProgress> = {}
  for (const [key, value] of Object.entries(source)) {
    result[key] = mapRootProgress(value)
  }
  return result
}

function mapProgressState(state: ProtoSyncProgressState): SyncProgress {
  return {
    status: mapSyncStatus(state.status),
    mode: mapSyncMode(state.mode),
    startedAt: int64ToNumber(state.startedAt),
    updatedAt: int64ToNumber(state.updatedAt),
    roots: state.roots.map((id) => int64ToNumber(id)),
    rootNames: state.rootNames,
    completedRoots: state.completedRoots.map((id) => int64ToNumber(id)),
    activeRoot:
      state.activeRoot != null ? int64ToNumber(state.activeRoot) : undefined,
    aggregateStats: mapCrawlStats(state.aggregateStats),
    rootProgress: mapRootProgressMap(state.rootProgress),
    catalogRoots: state.catalogRoots?.map((id) => int64ToNumber(id)),
    catalogRootNames: state.catalogRootNames,
    catalogRootProgress: state.catalogRootProgress
      ? mapRootProgressMap(state.catalogRootProgress)
      : undefined,
    incrementalStats: state.incrementalStats
      ? {
          changesFetched: int64ToNumber(state.incrementalStats.changesFetched),
          upserted: int64ToNumber(state.incrementalStats.upserted),
          deleted: int64ToNumber(state.incrementalStats.deleted),
          skippedUpserts: int64ToNumber(state.incrementalStats.skippedUpserts),
          skippedDeletes: int64ToNumber(state.incrementalStats.skippedDeletes),
          cursorBefore: int64ToNumber(state.incrementalStats.cursorBefore),
          cursorAfter: int64ToNumber(state.incrementalStats.cursorAfter),
        }
      : undefined,
    lastError: state.lastError,
    verification: state.verification
      ? {
          meiliDocCount: int64ToNumber(state.verification.meiliDocCount),
          crawledDocCount: int64ToNumber(state.verification.crawledDocCount),
          discoveredDocCount: int64ToNumber(
            state.verification.discoveredDocCount,
          ),
          skippedCount: int64ToNumber(state.verification.skippedCount),
          verified: state.verification.verified,
          warnings: state.verification.warnings,
        }
      : undefined,
    startedAtTs: timestampToProtoLike(state.startedAtTs),
    updatedAtTs: timestampToProtoLike(state.updatedAtTs),
  }
}

export function fromProtoSyncProgressState(
  state: ProtoSyncProgressState,
): SyncProgress {
  return mapProgressState(state)
}

export function fromProtoGetSyncProgressResponse(
  response: GetSyncProgressResponse,
): SyncProgress | null {
  if (!response.state) {
    return null
  }
  return mapProgressState(response.state)
}

export function fromProtoInspectRootsResponse(
  response: ProtoInspectRootsResponse,
): InspectRootsResponse {
  return {
    items: response.items.map((item) => ({
      folder_id: int64ToNumber(item.folderId),
      name: item.name,
      item_count: int64ToNumber(item.itemCount),
      estimated_total_docs: int64ToNumber(item.estimatedTotalDocs),
    })),
    errors: response.errors.map((item) => ({
      folder_id: int64ToNumber(item.folderId),
      message: item.message,
    })),
  }
}

export function toProtoSyncMode(mode: string): SyncMode {
  switch (mode) {
    case 'full':
      return SyncMode.FULL
    case 'incremental':
      return SyncMode.INCREMENTAL
    case 'auto':
    default:
      return SyncMode.AUTO
  }
}
