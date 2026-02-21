import { createLazyFileRoute } from '@tanstack/react-router'
import { AdminSyncPage } from '@/components/admin-sync-page'

export const Route = createLazyFileRoute('/admin/sync')({
  component: AdminSyncPage,
})
