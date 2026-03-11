<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors, formatBytes, formatUnit } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/snapshots'

const statusBadge = badgeColors({
  READY: badge.emerald,
  CREATING: badge.blue,
  UPLOADING: badge.blue,
  DELETING: badge.amber,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'snapshot_type', label: 'Type' },
  { key: 'disk_size_gb', label: 'Disk Size', format: 'number', transform: (v) => formatUnit(v, 'GB') },
  { key: 'storage_bytes', label: 'Storage', format: 'number', transform: (v) => formatBytes(v) },
  { key: 'architecture', label: 'Arch' },
  { key: 'auto_created', label: 'Auto', format: 'bool' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'snapshot_type', label: 'Type' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'status', label: 'Status' },
  { key: 'snapshot_type', label: 'Snapshot Type' },
  { key: 'disk_size_gb', label: 'Disk Size', transform: (v) => formatUnit(v, 'GB') },
  { key: 'storage_bytes', label: 'Storage', transform: (v) => formatBytes(v) },
  { key: 'architecture', label: 'Architecture' },
  { key: 'auto_created', label: 'Auto Created', format: 'bool' },
  { key: 'source_disk', label: 'Source Disk', transform: shortPath },
  { key: 'enable_confidential_compute', label: 'Confidential', format: 'bool' },
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
    export-filename="gcp-compute-snapshots.csv"
    title="Snapshots"
  />
</template>
