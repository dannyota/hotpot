<script setup lang="ts">
import { useRouter } from 'vue-router'
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const router = useRouter()
const ENDPOINT = '/api/v1/silver/inventory/software'

function decodeEntities(s: string): string {
  if (!s) return s
  return s
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
}

function goToMachine(machineId: string) {
  router.push({ path: '/silver/inventory/machines', query: { 'filter[resource_id]': machineId } })
}

const statusBadge = badgeColors({
  RUNNING: badge.emerald,
  STOPPED: badge.zinc,
  ERROR: badge.red,
})

const envBadge = badgeColors({
  PRODUCTION: badge.red,
  UAT: badge.amber,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold', transform: decodeEntities },
  { key: 'version', format: 'mono' },
  { key: 'publisher', transform: decodeEntities },
  { key: 'machine_hostname', label: 'Machine' },
  { key: 'machine_status', label: 'Status', badge: statusBadge },
  { key: 'machine_environment', label: 'Environment', badge: envBadge },
  { key: 'installed_on', label: 'Installed On', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'publisher' },
  { key: 'machine_status', label: 'Status' },
  { key: 'machine_environment', label: 'Environment' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'machine_hostname', label: 'Machine' },
  { key: 'machine_id', label: 'Machine ID', mono: true },
  { key: 'name', label: 'Name', transform: decodeEntities },
  { key: 'version', label: 'Version' },
  { key: 'publisher', label: 'Publisher', transform: decodeEntities },
  { key: 'machine_status', label: 'Status' },
  { key: 'machine_environment', label: 'Environment' },
  { key: 'installed_on', label: 'Installed On', format: 'date' },
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
    search-key="name"
    search-placeholder="Search by name..."
    export-filename="silver-inventory-software.csv"
    drawer-title-key="name"
    title="Software"
  >
    <template #cell-machine_hostname="{ value, row }">
      <button
        v-if="value"
        class="text-xs text-blue-600 dark:text-blue-400 hover:underline truncate max-w-[180px] block"
        :title="row.machine_id"
        @click.stop="goToMachine(row.machine_id)"
      >{{ value }}</button>
      <span v-else class="text-xs text-zinc-400">-</span>
    </template>
  </HTablePage>
</template>
