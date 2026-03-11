<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/compute/user-images'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
  ERROR: badge.red,
  SAVING: badge.blue,
  QUEUED: badge.amber,
  DEACTIVATED: badge.zinc,
})

function formatDisk(gb: any): string {
  const n = Number(gb)
  if (isNaN(n) || gb === null || gb === '') return String(gb ?? '')
  return `${n} GB`
}

function formatSize(bytes: any): string {
  const n = Number(bytes)
  if (isNaN(n) || bytes === null || bytes === '') return String(bytes ?? '')
  if (n >= 1073741824) return `${(n / 1073741824).toFixed(1)} GB`
  if (n >= 1048576) return `${(n / 1048576).toFixed(1)} MB`
  if (n >= 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${n} B`
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'min_disk', label: 'Min Disk' },
  { key: 'image_size', label: 'Image Size' },
  { key: 'region' },
  { key: 'created_at', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'min_disk', label: 'Min Disk', transform: (v) => formatDisk(v) },
  { key: 'image_size', label: 'Image Size', transform: (v) => formatSize(v) },
  { key: 'meta_data', label: 'Metadata' },
  { key: 'region', label: 'Region' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'created_at', label: 'Created', format: 'date' },
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
    export-filename="greennode-user-images.csv"
    title="User Images"
  >
    <template #cell-min_disk="{ value }">
      <span v-if="value != null" class="tabular-nums">{{ formatDisk(value) }}</span>
    </template>

    <template #cell-image_size="{ value }">
      <span v-if="value != null" class="tabular-nums">{{ formatSize(value) }}</span>
    </template>
  </HTablePage>
</template>
