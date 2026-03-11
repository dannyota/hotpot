<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors, formatUnit } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/disks'

const statusBadge = badgeColors({
  READY: badge.emerald,
  CREATING: badge.blue,
  FAILED: badge.red,
  DELETING: badge.amber,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'zone', transform: shortPath },
  { key: 'type', transform: shortPath, format: 'mono' },
  { key: 'size_gb', label: 'Size', format: 'number', transform: (v) => formatUnit(v, 'GB') },
  { key: 'architecture', label: 'Arch' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'zone' },
  { key: 'type' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'status', label: 'Status' },
  { key: 'zone', label: 'Zone', transform: shortPath },
  { key: 'type', label: 'Type', transform: shortPath },
  { key: 'size_gb', label: 'Size', transform: (v) => formatUnit(v, 'GB') },
  { key: 'architecture', label: 'Architecture' },
  { key: 'provisioned_iops', label: 'IOPS' },
  { key: 'provisioned_throughput', label: 'Throughput' },
  { key: 'source_image', label: 'Source Image', transform: shortPath },
  { key: 'source_snapshot', label: 'Source Snapshot', transform: shortPath },
  { key: 'enable_confidential_compute', label: 'Confidential', format: 'bool' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'self_link', label: 'Self Link', mono: true },
  { key: 'creation_timestamp', label: 'Created' },
  { key: 'last_attach_timestamp', label: 'Last Attach' },
  { key: 'last_detach_timestamp', label: 'Last Detach' },
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
    export-filename="gcp-compute-disks.csv"
    title="Persistent Disks"
  />
</template>
