<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/gold/lifecycle/software'

const eolStatusBadge = badgeColors({
  EOL_EXPIRED: badge.red,
  EOES_EXPIRED: badge.red,
  EOL_APPROACHING: badge.amber,
  SUPPORTED: badge.emerald,
})

const classificationBadge = badgeColors({
  OS: badge.blue,
  RUNTIME: badge.purple,
  DATABASE: badge.amber,
  FRAMEWORK: badge.zinc,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'version', format: 'mono' },
  { key: 'classification', badge: classificationBadge },
  { key: 'eol_status', label: 'EOL Status', badge: eolStatusBadge, transform: (v) => v?.replace(/_/g, ' ') ?? '' },
  { key: 'eol_date', label: 'EOL Date', format: 'date' },
  { key: 'eoes_date', label: 'EOES Date', format: 'date', sortable: false },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'eol_status', label: 'EOL Status' },
  { key: 'classification' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'version', label: 'Version' },
  { key: 'classification', label: 'Classification' },
  { key: 'eol_status', label: 'EOL Status', transform: (v) => v?.replace(/_/g, ' ') ?? '' },
  { key: 'eol_date', label: 'EOL Date', format: 'date' },
  { key: 'eoes_date', label: 'EOES Date', format: 'date' },
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
    search-placeholder="Search software..."
    export-filename="gold-lifecycle-software.csv"
    drawer-title-key="name"
    title="Software EOL"
  />
</template>
