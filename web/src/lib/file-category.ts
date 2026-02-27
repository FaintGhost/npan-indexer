export type SearchFilter = 'all' | 'doc' | 'image' | 'video' | 'archive' | 'other'

export const DEFAULT_FILTER: SearchFilter = 'all'

const FILTER_SET = new Set<SearchFilter>(['all', 'doc', 'image', 'video', 'archive', 'other'])

const DOC_EXTENSIONS = new Set([
  'pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx',
  'txt', 'md', 'markdown', 'rtf', 'odt', 'ods', 'odp',
  'csv', 'tsv', 'epub',
])

const IMAGE_EXTENSIONS = new Set([
  'jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg',
  'tif', 'tiff', 'heic', 'heif', 'avif', 'ico',
])

const VIDEO_EXTENSIONS = new Set([
  'mp4', 'mkv', 'mov', 'avi', 'wmv', 'flv', 'webm',
  'm4v', 'mpg', 'mpeg', 'ts', 'rmvb',
])

const MULTI_PART_ARCHIVE_EXTENSIONS = ['tar.gz', 'tar.bz2', 'tar.xz']

const ARCHIVE_EXTENSIONS = new Set([
  'zip', 'rar', '7z', 'tar', 'gz', 'tgz', 'bz2', 'xz', 'zst',
  ...MULTI_PART_ARCHIVE_EXTENSIONS,
])

export function normalizeSearchFilter(value: string | null | undefined): SearchFilter {
  if (!value) {
    return DEFAULT_FILTER
  }
  if (FILTER_SET.has(value as SearchFilter)) {
    return value as SearchFilter
  }
  return DEFAULT_FILTER
}

function getNormalizedExtension(fileName: string): string {
  const lowerName = fileName.trim().toLowerCase()
  if (!lowerName) {
    return ''
  }

  for (const ext of MULTI_PART_ARCHIVE_EXTENSIONS) {
    if (lowerName.endsWith(`.${ext}`)) {
      return ext
    }
  }

  const dotIndex = lowerName.lastIndexOf('.')
  if (dotIndex <= 0 || dotIndex === lowerName.length - 1) {
    return ''
  }

  return lowerName.slice(dotIndex + 1)
}

export function categorizeFileName(fileName: string): Exclude<SearchFilter, 'all'> {
  const ext = getNormalizedExtension(fileName)
  if (!ext) {
    return 'other'
  }

  if (DOC_EXTENSIONS.has(ext)) {
    return 'doc'
  }
  if (IMAGE_EXTENSIONS.has(ext)) {
    return 'image'
  }
  if (VIDEO_EXTENSIONS.has(ext)) {
    return 'video'
  }
  if (ARCHIVE_EXTENSIONS.has(ext)) {
    return 'archive'
  }

  return 'other'
}

export function matchesSearchFilter(fileName: string, filter: SearchFilter): boolean {
  if (filter === 'all') {
    return true
  }
  return categorizeFileName(fileName) === filter
}
