const KNOWN_EXTENSIONS = new Set([
  'pdf',
  'docx',
  'xlsx',
  'pptx',
  'doc',
  'xls',
  'ppt',
  'jpg',
  'jpeg',
  'png',
  'gif',
  'bmp',
  'mp4',
  'avi',
  'mov',
  'mkv',
  'zip',
  'rar',
  '7z',
  'tar',
  'gz',
  'exe',
  'apk',
  'bin',
  'iso',
  'dwg',
  'dxf',
  'cad',
  'txt',
  'csv',
  'json',
  'xml',
])

const VERSION_PREFIX_RE = /^[Vv](\d+\..+)$/

function reorderQuery(query: string): string {
  const words = query.split(/\s+/).filter((word) => word.length > 0)
  if (words.length === 0) {
    return query
  }

  const extensions: string[] = []
  const terms: string[] = []

  for (const word of words) {
    if (KNOWN_EXTENSIONS.has(word.toLowerCase())) {
      extensions.push(word)
      continue
    }
    terms.push(word)
  }

  if (extensions.length === 0 || terms.length === 0) {
    return query
  }

  return [...extensions, ...terms].join(' ')
}

export function normalizeSearchQuery(query: string): string {
  const words = query.split(/\s+/).filter((word) => word.length > 0)
  if (words.length === 0) {
    return query
  }

  const expanded: string[] = []

  for (const word of words) {
    const dotIndex = word.lastIndexOf('.')
    if (dotIndex > 0 && dotIndex < word.length - 1) {
      const extension = word.slice(dotIndex + 1)
      if (KNOWN_EXTENSIONS.has(extension.toLowerCase())) {
        expanded.push(word.slice(0, dotIndex), extension)
        continue
      }
    }

    expanded.push(word)
  }

  const normalized = expanded.map((word) => {
    const match = VERSION_PREFIX_RE.exec(word)
    return match?.[1] ?? word
  })

  return reorderQuery(normalized.join(' '))
}
