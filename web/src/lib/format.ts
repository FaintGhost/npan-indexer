export function formatBytes(size: number): string {
  const value = Number(size || 0)
  if (!value) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let idx = 0
  let num = value
  while (num >= 1024 && idx < units.length - 1) {
    num /= 1024
    idx += 1
  }
  const decimals = num >= 10 || idx === 0 ? 0 : (num % 1 === 0 ? 0 : 1)
  return num.toFixed(decimals) + ' ' + units[idx]
}

export function formatTime(ts: number): string {
  const raw = Number(ts || 0)
  if (!raw) return '-'
  const millis = raw > 1_000_000_000_000 ? raw : raw * 1000
  const d = new Date(millis)
  if (Number.isNaN(d.getTime())) return '-'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${m}-${day} ${hh}:${mm}`
}
