<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/compute/os-images'

const licenceBadge = badgeColors({
  FREE: badge.emerald,
}, badge.amber)

function formatMemory(mb: any): string {
  const n = Number(mb)
  if (isNaN(n) || mb === null || mb === '') return String(mb ?? '')
  if (n >= 1024) return `${(n / 1024).toFixed(n % 1024 === 0 ? 0 : 1)} GB`
  return `${n} MB`
}

function formatDisk(gb: any): string {
  const n = Number(gb)
  if (isNaN(n) || gb === null || gb === '') return String(gb ?? '')
  return `${n} GB`
}

const columns: ColumnDef[] = [
  { key: 'image_type', label: 'Type', format: 'bold' },
  { key: 'image_version', label: 'Version' },
  { key: 'description', label: 'Description' },
  { key: 'licence', label: 'Licence', badge: licenceBadge },
  { key: 'package_limit_cpu', label: 'CPU Limit', format: 'number' },
  { key: 'package_limit_memory', label: 'Memory Limit', format: 'number' },
  { key: 'package_limit_disk_size', label: 'Disk Limit', format: 'number' },
  { key: 'region' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'image_type', label: 'Type' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'image_type', label: 'Type' },
  { key: 'image_version', label: 'Version' },
  { key: 'description', label: 'Description' },
  { key: 'licence', label: 'Licence' },
  { key: 'license_key', label: 'License Key', mono: true },
  { key: 'package_limit_cpu', label: 'CPU Limit' },
  { key: 'package_limit_memory', label: 'Memory Limit', transform: (v) => formatMemory(v) },
  { key: 'package_limit_disk_size', label: 'Disk Limit', transform: (v) => formatDisk(v) },
  { key: 'zone_id', label: 'Zone' },
  { key: 'region', label: 'Region' },
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
    search-key="description"
    search-placeholder="Search by description..."
    export-filename="greennode-os-images.csv"
    drawer-title-key="image_type"
    title="OS Images"
  >
    <template #cell-package_limit_cpu="{ value }">
      <span v-if="value" class="tabular-nums">{{ value }} vCPU</span>
    </template>

    <template #cell-package_limit_memory="{ value }">
      <span v-if="value" class="tabular-nums">{{ formatMemory(value) }}</span>
    </template>

    <template #cell-package_limit_disk_size="{ value }">
      <span v-if="value" class="tabular-nums">{{ formatDisk(value) }}</span>
    </template>
  </HTablePage>
</template>
