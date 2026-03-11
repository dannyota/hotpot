<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/silver/inventory/k8s-nodes'

const statusBadge = badgeColors({
  RUNNING: badge.emerald,
  READY: badge.emerald,
  STOPPED: badge.red,
  NOT_READY: badge.red,
  NOTREADY: badge.red,
})

const provisioningBadge = badgeColors({
  SPOT: badge.amber,
  STANDARD: badge.blue,
})

const columns: ColumnDef[] = [
  { key: 'node_name', label: 'Node', format: 'bold' },
  { key: 'cluster_name', label: 'Cluster' },
  { key: 'node_pool', label: 'Pool' },
  { key: 'status', badge: statusBadge },
  { key: 'provisioning', badge: provisioningBadge },
  { key: 'cloud_project', label: 'Project' },
  { key: 'cloud_zone', label: 'Zone' },
  { key: 'internal_ip', label: 'Internal IP', format: 'mono' },
  { key: 'external_ip', label: 'External IP', format: 'mono' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
  { key: 'normalized_at', format: 'date' },
]

const filters: FilterDef[] = [
  { key: 'cluster_name', label: 'Cluster' },
  { key: 'status' },
  { key: 'cloud_project', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'node_name', label: 'Node' },
  { key: 'cluster_name', label: 'Cluster' },
  { key: 'node_pool', label: 'Pool' },
  { key: 'status', label: 'Status' },
  { key: 'provisioning', label: 'Provisioning' },
  { key: 'cloud_project', label: 'Project' },
  { key: 'cloud_zone', label: 'Zone' },
  { key: 'cloud_machine_type', label: 'Machine Type' },
  { key: 'internal_ip', label: 'Internal IP', mono: true },
  { key: 'external_ip', label: 'External IP', mono: true },
  { key: 'first_collected_at', label: 'First Seen', format: 'date' },
  { key: 'collected_at', label: 'Last Seen', format: 'date' },
  { key: 'normalized_at', label: 'Normalized', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-collected_at"
    search-key="node_name"
    search-placeholder="Search by node name..."
    export-filename="silver-k8s-nodes.csv"
    drawer-title-key="node_name"
    title="K8s Nodes"
  />
</template>
