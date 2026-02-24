import type { PropsWithChildren } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TransportProvider } from '@connectrpc/connect-query'
import { createNpanTransport } from '@/lib/connect-transport'

export function createTestProvider(headers?: Record<string, string>) {
  const client = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  })
  const transport = createNpanTransport(headers)

  return function TestProvider({ children }: PropsWithChildren) {
    return (
      <TransportProvider transport={transport}>
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      </TransportProvider>
    )
  }
}
