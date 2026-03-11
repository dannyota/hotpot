<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/target-pools'

const affinityBadge = badgeColors({
  CLIENT_IP: badge.blue,
  NONE: badge.zinc,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'region', transform: shortPath },
  { key: 'session_affinity', label: 'Session Affinity', badge: affinityBadge },
  { key: 'failover_ratio', label: 'Failover Ratio', format: 'number' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'region' },
  { key: 'session_affinity', label: 'Affinity' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'region', label: 'Region', transform: shortPath },
  { key: 'session_affinity', label: 'Session Affinity' },
  { key: 'failover_ratio', label: 'Failover Ratio' },
  { key: 'backup_pool', label: 'Backup Pool', transform: shortPath },
  { key: 'security_policy', label: 'Security Policy' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'self_link', label: 'Self Link', mono: true },
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
    export-filename="gcp-compute-target-pools.csv"
    title="Target Pools"
  />
</template>
