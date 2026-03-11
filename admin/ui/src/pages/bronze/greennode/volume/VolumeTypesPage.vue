<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/volume/volume-types'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'iops', label: 'IOPS', format: 'number' },
  { key: 'max_size', label: 'Max Size', format: 'number' },
  { key: 'min_size', label: 'Min Size', format: 'number' },
  { key: 'through_put', label: 'Throughput', format: 'number' },
  { key: 'zone_id', label: 'Zone ID', format: 'mono' },
  { key: 'region' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'region' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'iops', label: 'IOPS' },
  { key: 'max_size', label: 'Max Size' },
  { key: 'min_size', label: 'Min Size' },
  { key: 'through_put', label: 'Throughput' },
  { key: 'zone_id', label: 'Zone ID' },
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
    search-key="q"
    export-filename="greennode-volume-types.csv"
    title="Volume Types"
  />
</template>
