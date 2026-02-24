import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { QueryClientProvider } from '@tanstack/react-query'
import { TransportProvider } from '@connectrpc/connect-query'
import { routeTree } from './routeTree.gen'
import { appQueryClient, appTransport } from '@/lib/connect-transport'
import './app.css'

const router = createRouter({
  routeTree,
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const root = document.getElementById('root')
if (root) {
  createRoot(root).render(
    <StrictMode>
      <TransportProvider transport={appTransport}>
        <QueryClientProvider client={appQueryClient}>
          <RouterProvider router={router} />
        </QueryClientProvider>
      </TransportProvider>
    </StrictMode>,
  )
}
