import { createLazyFileRoute } from '@tanstack/react-router'
import { AdminSyncPage } from '@/components/admin-sync-page'

export const Route = createLazyFileRoute('/_admin/sync')({
  component: AdminSyncPage,
})
