import { useState, useCallback } from 'react'
import { Code, ConnectError } from '@connectrpc/connect'
import { callUnaryMethod } from '@connectrpc/connect-query-core'
import { getSyncProgress as getSyncProgressMethod } from '@/gen/npan/v1/api-AdminService_connectquery'
import {
  ADMIN_API_KEY_STORAGE_KEY as STORAGE_KEY,
  createNpanTransport,
} from '@/lib/connect-transport'

export function useAdminAuth() {
  const [apiKey, setApiKey] = useState<string | null>(
    () => localStorage.getItem(STORAGE_KEY),
  )
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const needsAuth = apiKey === null

  const validate = useCallback(async (key: string): Promise<boolean> => {
    if (!key.trim()) {
      setError('请输入 API Key')
      return false
    }

    setLoading(true)
    setError(null)

    try {
      const transport = createNpanTransport({ 'X-API-Key': key })
      await callUnaryMethod(transport, getSyncProgressMethod, {})
      localStorage.setItem(STORAGE_KEY, key)
      setApiKey(key)
      setLoading(false)
      return true
    } catch (err) {
      setLoading(false)
      if (err instanceof ConnectError && err.code === Code.NotFound) {
        // 404 means auth passed but no sync progress yet — treat as success
        localStorage.setItem(STORAGE_KEY, key)
        setApiKey(key)
        return true
      }
      if (err instanceof ConnectError && err.code === Code.Unauthenticated) {
        setError('API Key 无效')
      } else {
        setError(err instanceof Error ? err.message : '验证失败')
      }
      return false
    }
  }, [])

  const on401 = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY)
    setApiKey(null)
    setError(null)
  }, [])

  const getHeaders = useCallback((): Record<string, string> => {
    return apiKey ? { 'X-API-Key': apiKey } : {}
  }, [apiKey])

  return { needsAuth, apiKey, error, loading, validate, on401, getHeaders }
}
