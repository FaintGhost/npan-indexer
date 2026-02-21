export interface FileIconInfo {
  category: 'archive' | 'installer' | 'firmware' | 'document' | 'default'
  bg: string
  text: string
}

export function getFileIcon(filename: string): FileIconInfo {
  const extMatch = filename.match(/\.([a-zA-Z0-9]+)$/)
  const ext = extMatch?.[1]?.toLowerCase() ?? ''
  const nameLower = filename.toLowerCase()

  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
    return { category: 'archive', bg: 'bg-amber-100', text: 'text-amber-600' }
  }
  if (['apk', 'ipa', 'exe', 'dmg'].includes(ext) || nameLower.includes('安装包')) {
    return { category: 'installer', bg: 'bg-emerald-100', text: 'text-emerald-600' }
  }
  if (['bin', 'iso', 'img', 'rom'].includes(ext) || nameLower.includes('固件')) {
    return { category: 'firmware', bg: 'bg-purple-100', text: 'text-purple-600' }
  }
  if (['pdf', 'doc', 'docx', 'txt', 'md'].includes(ext)) {
    return { category: 'document', bg: 'bg-rose-100', text: 'text-rose-500' }
  }
  return { category: 'default', bg: 'bg-blue-50', text: 'text-blue-500' }
}
