import { useState, useCallback } from 'react'
import { apiGet, ApiError } from '@/lib/api-client'
import { SyncProgressSchema } from '@/lib/sync-schemas'

const STORAGE_KEY = 'npan_admin_api_key'

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
      await apiGet(
        '/api/v1/admin/sync',
        {},
        SyncProgressSchema,
        { headers: { 'X-API-Key': key } },
      )
      localStorage.setItem(STORAGE_KEY, key)
      setApiKey(key)
      setLoading(false)
      return true
    } catch (err) {
      setLoading(false)
      if (err instanceof ApiError && err.status === 401) {
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
