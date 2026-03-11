<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/endpoint-apps'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'version' },
  { key: 'publisher' },
  { key: 'agent_id', label: 'Agent ID', format: 'mono' },
  { key: 'size', format: 'number' },
  { key: 'installed_date', label: 'Installed', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'publisher' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'version', label: 'Version' },
  { key: 'publisher', label: 'Publisher' },
  { key: 'agent_id', label: 'Agent ID', mono: true },
  { key: 'size', label: 'Size', transform: (v) => v != null ? String(v) : '' },
  { key: 'installed_date', label: 'Installed', format: 'date' },
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
    search-placeholder="Search by name..."
    export-filename="s1-endpoint-apps.csv"
    drawer-title-key="name"
    title="Endpoint Apps"
  />
</template>
