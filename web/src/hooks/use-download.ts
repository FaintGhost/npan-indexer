import { useState, useRef, useCallback } from 'react'
import { apiGet } from '@/lib/api-client'
import { DownloadURLResponseSchema } from '@/lib/schemas'

type DownloadStatus = 'idle' | 'loading' | 'success' | 'error'

export function useDownload() {
  const [statuses, setStatuses] = useState<Map<number, DownloadStatus>>(new Map())
  const cacheRef = useRef<Map<number, string>>(new Map())

  const setStatus = useCallback((fileId: number, status: DownloadStatus) => {
    setStatuses((prev) => {
      const next = new Map(prev)
      next.set(fileId, status)
      return next
    })
  }, [])

  const download = useCallback(async (fileId: number) => {
    // Check cache first
    const cached = cacheRef.current.get(fileId)
    if (cached) {
      setStatus(fileId, 'success')
      window.open(cached, '_blank', 'noopener,noreferrer')
      setTimeout(() => setStatus(fileId, 'idle'), 1500)
      return
    }

    setStatus(fileId, 'loading')

    try {
      const result = await apiGet(
        '/api/v1/app/download-url',
        { file_id: fileId },
        DownloadURLResponseSchema,
      )

      cacheRef.current.set(fileId, result.download_url)
      setStatus(fileId, 'success')
      window.open(result.download_url, '_blank', 'noopener,noreferrer')
      setTimeout(() => setStatus(fileId, 'idle'), 1500)
    } catch {
      setStatus(fileId, 'error')
    }
  }, [setStatus])

  const getStatus = useCallback(
    (fileId: number): DownloadStatus => statuses.get(fileId) ?? 'idle',
    [statuses],
  )

  return { download, getStatus }
}
