<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/app-inventory'

const columns: ColumnDef[] = [
  { key: 'application_name', label: 'Name', format: 'bold' },
  { key: 'application_vendor', label: 'Vendor' },
  { key: 'endpoints_count', label: 'Endpoints', format: 'number' },
  { key: 'application_versions_count', label: 'Versions', format: 'number' },
  { key: 'estimate', label: 'Estimate', format: 'bool' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'application_vendor', label: 'Vendor' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'application_name', label: 'Name' },
  { key: 'application_vendor', label: 'Vendor' },
  { key: 'endpoints_count', label: 'Endpoints', transform: (v) => v != null ? String(v) : '' },
  { key: 'application_versions_count', label: 'Versions', transform: (v) => v != null ? String(v) : '' },
  { key: 'estimate', label: 'Estimate', format: 'bool' },
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
    search-key="application_name"
    search-placeholder="Search by name..."
    export-filename="s1-app-inventory.csv"
    drawer-title-key="application_name"
    title="App Inventory"
  />
</template>
