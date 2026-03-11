<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/volume/snapshots'

const statusBadge = badgeColors({
  AVAILABLE: badge.emerald,
  'IN-USE': badge.blue,
  ERROR: badge.red,
  CREATING: badge.amber,
  DELETING: badge.amber,
})

function formatSize(size: any): string {
  const n = Number(size)
  if (!isNaN(n) && size !== null && size !== '') return `${n} GB`
  return String(size ?? '')
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'snapshot_id', label: 'Snapshot ID', format: 'mono' },
  { key: 'status', badge: statusBadge },
  { key: 'size' },
  { key: 'volume_size', label: 'Volume Size' },
  { key: 'created_at_api', label: 'Created', format: 'date' },
]

const filters: FilterDef[] = [
  { key: 'status' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'snapshot_id', label: 'Snapshot ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'size', label: 'Size', transform: (v) => formatSize(v) },
  { key: 'volume_size', label: 'Volume Size', transform: (v) => formatSize(v) },
  { key: 'created_at_api', label: 'Created', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-id"
    search-key="q"
    export-filename="greennode-volume-snapshots.csv"
    title="Volume Snapshots"
  >
    <template #cell-size="{ value }">
      <span class="tabular-nums">{{ formatSize(value) }}</span>
    </template>

    <template #cell-volume_size="{ value }">
      <span class="tabular-nums">{{ formatSize(value) }}</span>
    </template>
  </HTablePage>
</template>
