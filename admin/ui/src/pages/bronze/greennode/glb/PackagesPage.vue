<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/glb/packages'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'description' },
  { key: 'enabled', format: 'bool' },
  { key: 'base_connection_rate', label: 'Connection Rate', format: 'number' },
  { key: 'base_domestic_traffic_total', label: 'Domestic Traffic', format: 'number' },
  { key: 'base_non_domestic_traffic_total', label: 'Intl Traffic', format: 'number' },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'description_en', label: 'Description (EN)' },
  { key: 'enabled', label: 'Enabled', format: 'bool' },
  { key: 'base_sku', label: 'Base SKU', mono: true },
  { key: 'base_connection_rate', label: 'Connection Rate' },
  { key: 'base_domestic_traffic_total', label: 'Domestic Traffic Total' },
  { key: 'base_non_domestic_traffic_total', label: 'Intl Traffic Total' },
  { key: 'connection_sku', label: 'Connection SKU', mono: true },
  { key: 'domestic_traffic_sku', label: 'Domestic Traffic SKU', mono: true },
  { key: 'non_domestic_traffic_sku', label: 'Intl Traffic SKU', mono: true },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'updated_at_api', label: 'Updated', format: 'date' },
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
    export-filename="greennode-glb-packages.csv"
    title="GLB Packages"
  />
</template>
