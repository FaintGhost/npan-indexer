import { useState, useRef, useCallback } from 'react'
import { callUnaryMethod } from '@connectrpc/connect-query-core'
import { appDownloadURL as appDownloadURLMethod } from '@/gen/npan/v1/api-AppService_connectquery'
import { appTransport } from '@/lib/connect-transport'
import { fromProtoAppDownloadURLResponse } from '@/lib/connect-app-adapter'

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
      const result = await callUnaryMethod(
        appTransport,
        appDownloadURLMethod,
        { fileId: BigInt(fileId) },
      )
      const downloadURL = fromProtoAppDownloadURLResponse(result)

      cacheRef.current.set(fileId, downloadURL)
      setStatus(fileId, 'success')
      window.open(downloadURL, '_blank', 'noopener,noreferrer')
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
