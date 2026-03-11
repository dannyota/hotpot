<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors, formatBytes, formatUnit } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/images'

const statusBadge = badgeColors({
  READY: badge.emerald,
  PENDING: badge.blue,
  FAILED: badge.red,
  DELETING: badge.amber,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'family' },
  { key: 'architecture', label: 'Arch' },
  { key: 'disk_size_gb', label: 'Disk Size', format: 'number', transform: (v) => formatUnit(v, 'GB') },
  { key: 'source_type', label: 'Source Type' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'family' },
  { key: 'architecture' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'status', label: 'Status' },
  { key: 'family', label: 'Family' },
  { key: 'architecture', label: 'Architecture' },
  { key: 'disk_size_gb', label: 'Disk Size', transform: (v) => formatUnit(v, 'GB') },
  { key: 'archive_size_bytes', label: 'Archive Size', transform: (v) => formatBytes(v) },
  { key: 'source_type', label: 'Source Type' },
  { key: 'source_disk', label: 'Source Disk', transform: shortPath },
  { key: 'source_image', label: 'Source Image', transform: shortPath },
  { key: 'source_snapshot', label: 'Source Snapshot', transform: shortPath },
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
    export-filename="gcp-compute-images.csv"
    title="Compute Images"
  />
</template>
