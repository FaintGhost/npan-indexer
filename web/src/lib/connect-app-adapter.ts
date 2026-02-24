import { ItemType } from '@/gen/npan/v1/api_pb'
import type {
  AppSearchResponse,
  IndexDocument as ProtoIndexDocument,
} from '@/gen/npan/v1/api_pb'
import type { IndexDocument, SearchResponse } from '@/lib/schemas'

function int64ToNumber(value: bigint): number {
  return Number(value)
}

function mapItemType(value: ItemType): IndexDocument['type'] {
  switch (value) {
    case ItemType.FOLDER:
      return 'folder'
    case ItemType.FILE:
    case ItemType.UNSPECIFIED:
    default:
      return 'file'
  }
}

function mapIndexDocument(item: ProtoIndexDocument): IndexDocument {
  return {
    doc_id: item.docId,
    source_id: int64ToNumber(item.sourceId),
    type: mapItemType(item.type),
    name: item.name,
    path_text: item.pathText,
    parent_id: int64ToNumber(item.parentId),
    modified_at: int64ToNumber(item.modifiedAt),
    created_at: int64ToNumber(item.createdAt),
    size: int64ToNumber(item.size),
    sha1: item.sha1,
    in_trash: item.inTrash,
    is_deleted: item.isDeleted,
    highlighted_name: item.highlightedName,
  }
}

export function fromProtoAppSearchResponse(response: AppSearchResponse): SearchResponse {
  const result = response.result
  if (!result) {
    return {
      items: [],
      total: 0,
    }
  }

  return {
    items: result.items.map(mapIndexDocument),
    total: int64ToNumber(result.total),
  }
}
