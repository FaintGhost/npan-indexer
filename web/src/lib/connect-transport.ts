import type { Interceptor, Transport } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { QueryClient } from '@tanstack/react-query'

export const ADMIN_API_KEY_STORAGE_KEY = 'npan_admin_api_key'

function createAuthInterceptor(
  explicitHeaders?: Record<string, string>,
): Interceptor {
  return (next) => async (req) => {
    const apiKey = localStorage.getItem(ADMIN_API_KEY_STORAGE_KEY)
    if (apiKey && !req.header.has('X-API-Key')) {
      req.header.set('X-API-Key', apiKey)
    }

    if (explicitHeaders) {
      for (const [key, value] of Object.entries(explicitHeaders)) {
        req.header.set(key, value)
      }
    }

    return next(req)
  }
}

export function createNpanTransport(
  explicitHeaders?: Record<string, string>,
): Transport {
  return createConnectTransport({
    baseUrl: '/',
    useBinaryFormat: false,
    interceptors: [createAuthInterceptor(explicitHeaders)],
  })
}

export const appTransport = createNpanTransport()

export const appQueryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: false,
    },
  },
})
