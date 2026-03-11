<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/container/node-pools'

const statusBadge = badgeColors({
  RUNNING: badge.emerald,
  PROVISIONING: badge.blue,
  RECONCILING: badge.amber,
  STOPPING: badge.amber,
  ERROR: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'version', format: 'mono' },
  { key: 'initial_node_count', label: 'Nodes', format: 'number' },
  { key: 'cluster_name', label: 'Cluster' },
  { key: 'cluster_location', label: 'Location', format: 'mono' },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'status_message', label: 'Status Message' },
  { key: 'version', label: 'Version' },
  { key: 'initial_node_count', label: 'Initial Node Count' },
  { key: 'pod_ipv4_cidr_size', label: 'Pod CIDR Size' },
  { key: 'cluster_name', label: 'Cluster' },
  { key: 'cluster_location', label: 'Location' },
  { key: 'project_id', label: 'Project ID', mono: true },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-id"
    search-placeholder="Search by name..."
    export-filename="gcp-gke-node-pools.csv"
    drawer-title-key="name"
    title="GKE Node Pools"
  />
</template>
