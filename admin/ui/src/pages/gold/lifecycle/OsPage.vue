<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/gold/lifecycle/os'

const eolStatusBadge = badgeColors({
  ACTIVE: badge.emerald,
  EOAS_EXPIRED: badge.amber,
  EOL_EXPIRED: badge.red,
  EOES_EXPIRED: badge.red,
  UNKNOWN: badge.zinc,
})

const columns: ColumnDef[] = [
  { key: 'hostname', format: 'bold' },
  { key: 'os_type', label: 'OS Type', badge: () => badge.zinc },
  { key: 'os_name', label: 'OS' },
  { key: 'eol_status', label: 'EOL Status', badge: eolStatusBadge },
  { key: 'eol_product_name', label: 'Product' },
  { key: 'eol_cycle', label: 'Cycle' },
  { key: 'eol_date', label: 'EOL Date', format: 'date' },
  { key: 'eoas_date', label: 'EOAS Date', format: 'date' },
  { key: 'latest_version', label: 'Latest' },
  { key: 'first_detected_at', label: 'First Detected', format: 'date' },
  { key: 'detected_at', label: 'Detected', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'eol_status', label: 'EOL Status' },
  { key: 'os_type', label: 'OS Type' },
  { key: 'eol_product_name', label: 'Product' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'machine_id', label: 'Machine ID', mono: true },
  { key: 'hostname', label: 'Hostname' },
  { key: 'os_type', label: 'OS Type' },
  { key: 'os_name', label: 'OS Name' },
  { key: 'eol_status', label: 'EOL Status' },
  { key: 'eol_product_slug', label: 'EOL Product Slug' },
  { key: 'eol_product_name', label: 'EOL Product Name' },
  { key: 'eol_cycle', label: 'EOL Cycle' },
  { key: 'eol_date', label: 'EOL Date', format: 'date' },
  { key: 'eoas_date', label: 'EOAS Date', format: 'date' },
  { key: 'eoes_date', label: 'EOES Date', format: 'date' },
  { key: 'latest_version', label: 'Latest Version' },
  { key: 'detected_at', label: 'Detected', format: 'date' },
  { key: 'first_detected_at', label: 'First Detected', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-detected_at"
    search-key="q"
    search-placeholder="Search by hostname..."
    export-filename="gold-lifecycle-os.csv"
    drawer-title-key="hostname"
    title="OS EOL"
  />
</template>
