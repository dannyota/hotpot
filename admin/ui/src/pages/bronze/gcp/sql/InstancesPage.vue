<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/sql/instances'

const stateBadge = badgeColors({
  RUNNABLE: badge.emerald,
  SUSPENDED: badge.amber,
  PENDING_DELETE: badge.red,
})

function dbVersionBadge(ver: string): string {
  if (!ver) return badge.zinc
  if (ver.startsWith('POSTGRES')) return badge.blue
  if (ver.startsWith('MYSQL')) return badge.amber
  if (ver.startsWith('SQLSERVER')) return badge.purple
  return badge.zinc
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'database_version', label: 'Version', badge: dbVersionBadge },
  { key: 'state', badge: stateBadge },
  { key: 'instance_type', label: 'Type' },
  { key: 'region' },
  { key: 'gce_zone', label: 'Zone' },
  { key: 'connection_name', label: 'Connection', format: 'mono' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'state' },
  { key: 'database_version', label: 'Version' },
  { key: 'region' },
  { key: 'instance_type', label: 'Type' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'database_version', label: 'Database Version' },
  { key: 'state', label: 'State' },
  { key: 'instance_type', label: 'Instance Type' },
  { key: 'region', label: 'Region' },
  { key: 'gce_zone', label: 'Zone' },
  { key: 'secondary_gce_zone', label: 'Secondary Zone' },
  { key: 'connection_name', label: 'Connection Name', mono: true },
  { key: 'service_account_email_address', label: 'Service Account', mono: true },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'self_link', label: 'Self Link', mono: true },
  { key: 'first_collected_at', label: 'First Seen', format: 'date' },
  { key: 'collected_at', label: 'Last Seen', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-collected_at"
    search-placeholder="Search by name..."
    export-filename="gcp-sql-instances.csv"
    drawer-title-key="name"
    title="Cloud SQL Instances"
  />
</template>
