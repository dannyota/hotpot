<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/container/clusters'

const statusBadge = badgeColors({
  RUNNING: badge.emerald,
  PROVISIONING: badge.blue,
  RECONCILING: badge.amber,
  ERROR: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'location', format: 'mono' },
  { key: 'current_master_version', label: 'Master Version', format: 'mono' },
  { key: 'current_node_version', label: 'Node Version', format: 'mono' },
  { key: 'current_node_count', label: 'Nodes', format: 'number' },
  { key: 'network', format: 'mono' },
  { key: 'endpoint', format: 'mono' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'location' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'location', label: 'Location' },
  { key: 'current_master_version', label: 'Master Version' },
  { key: 'current_node_version', label: 'Node Version' },
  { key: 'current_node_count', label: 'Node Count', transform: (v) => v != null ? String(v) : '' },
  { key: 'network', label: 'Network' },
  { key: 'subnetwork', label: 'Subnetwork' },
  { key: 'endpoint', label: 'Endpoint', mono: true },
  { key: 'project_id', label: 'Project ID', mono: true },
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
    export-filename="gcp-gke-clusters.csv"
    drawer-title-key="name"
    title="GKE Clusters"
  />
</template>
