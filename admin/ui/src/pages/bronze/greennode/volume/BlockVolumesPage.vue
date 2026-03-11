<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import HJsonPeek from '@/components/app/HJsonPeek.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/volume/block-volumes'

const statusBadge = badgeColors({
  'IN-USE': badge.emerald,
  AVAILABLE: badge.blue,
  ERROR: badge.red,
})

function formatSize(size: any): string {
  const n = Number(size)
  if (!isNaN(n) && size !== null && size !== '') return `${n} GB`
  return String(size ?? '')
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'size' },
  { key: 'region' },
  { key: 'persistent_volume', label: 'Persistent', format: 'bool' },
  { key: 'multi_attach', label: 'Multi-Attach', format: 'bool' },
  { key: 'attached_machine_json', label: 'Attached Machine', sortable: false },
  { key: 'created_at_api', label: 'Created', format: 'date' },
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
  { key: 'size', label: 'Size', transform: (v) => formatSize(v) },
  { key: 'region', label: 'Region' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'persistent_volume', label: 'Persistent Volume', format: 'bool' },
  { key: 'multi_attach', label: 'Multi-Attach', format: 'bool' },
  { key: 'attached_machine_json', label: 'Attached Machine' },
  { key: 'created_at_api', label: 'Created', format: 'date' },
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
    search-key="q"
    export-filename="greennode-block-volumes.csv"
    title="Block Volumes"
  >
    <template #cell-size="{ value }">
      <span class="tabular-nums">{{ formatSize(value) }}</span>
    </template>

    <template #cell-attached_machine_json="{ value, row, showJson }">
      <HJsonPeek
        :data="value"
        :title="`Attached Machine — ${row.name}`"
        @click="showJson"
      />
    </template>
  </HTablePage>
</template>
