import DOMPurify from 'dompurify'
import type { IndexDocument } from '@/lib/schemas'

interface HighlightField {
  value?: string
}

interface HighlightResult {
  name?: HighlightField
}

const HIGHLIGHT_SANITIZE_OPTIONS = {
  ALLOWED_TAGS: ['mark'],
  ALLOWED_ATTR: [],
  ALLOW_DATA_ATTR: false,
  ALLOW_ARIA_ATTR: false,
}

export interface MeiliHit {
  doc_id: string
  source_id: number | string
  type: 'file' | 'folder' | string
  name: string
  path_text: string
  parent_id: number | string
  modified_at: number | string
  created_at: number | string
  size: number | string
  sha1: string
  in_trash: boolean
  is_deleted: boolean
  file_category?: IndexDocument['file_category']
  _highlightResult?: HighlightResult
}

function toNumber(value: number | string | undefined): number {
  if (typeof value === 'number') {
    return value
  }

  if (typeof value === 'string' && value.trim() !== '') {
    return Number(value)
  }

  return 0
}

function toItemType(value: MeiliHit['type']): IndexDocument['type'] {
  return value === 'folder' ? 'folder' : 'file'
}

function getHighlightedName(hit: MeiliHit): string | undefined {
  const highlightedName = hit._highlightResult?.name?.value
  if (!highlightedName) {
    return undefined
  }

  const sanitizedHighlight = DOMPurify.sanitize(
    highlightedName,
    HIGHLIGHT_SANITIZE_OPTIONS,
  ).trim()

  return sanitizedHighlight || undefined
}

export function fromMeiliHit(hit: MeiliHit): IndexDocument {
  return {
    doc_id: hit.doc_id,
    source_id: toNumber(hit.source_id),
    type: toItemType(hit.type),
    name: hit.name,
    path_text: hit.path_text,
    parent_id: toNumber(hit.parent_id),
    modified_at: toNumber(hit.modified_at),
    created_at: toNumber(hit.created_at),
    size: toNumber(hit.size),
    sha1: hit.sha1,
    in_trash: hit.in_trash,
    is_deleted: hit.is_deleted,
    file_category: hit.file_category,
    highlighted_name: getHighlightedName(hit),
  }
}
