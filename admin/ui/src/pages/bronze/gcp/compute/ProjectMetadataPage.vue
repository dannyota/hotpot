<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/project-metadata'

const tierBadge = badgeColors({
  PREMIUM: badge.emerald,
  STANDARD: badge.blue,
})

const xpnBadge = badgeColors({
  HOST: badge.purple,
  UNSPECIFIED_XPN_PROJECT_STATUS: badge.zinc,
})

function shortAccount(v: string): string {
  if (!v) return ''
  return v.replace(/@.*/, '')
}

function xpnLabel(v: string): string {
  if (v === 'UNSPECIFIED_XPN_PROJECT_STATUS') return 'UNSPECIFIED'
  return v?.replace(/_/g, ' ') ?? ''
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'default_service_account', label: 'Service Account', format: 'mono', transform: shortAccount },
  { key: 'default_network_tier', label: 'Network Tier', badge: tierBadge },
  { key: 'xpn_project_status', label: 'XPN Status', badge: xpnBadge, transform: xpnLabel },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'default_network_tier', label: 'Network Tier' },
  { key: 'xpn_project_status', label: 'XPN Status' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'default_service_account', label: 'Default Service Account', mono: true },
  { key: 'default_network_tier', label: 'Network Tier' },
  { key: 'xpn_project_status', label: 'XPN Status' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'creation_timestamp', label: 'Created' },
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
    default-sort="-creation_timestamp"
    export-filename="gcp-compute-project-metadata.csv"
    title="Project Metadata"
  />
</template>
