<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/loadbalancer/packages'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'type' },
  { key: 'connection_number', label: 'Connections', format: 'number' },
  { key: 'data_transfer', label: 'Data Transfer', format: 'number' },
  { key: 'mode' },
  { key: 'lb_type', label: 'LB Type' },
  { key: 'display_lb_type', label: 'Display LB Type' },
  { key: 'region' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'type' },
  { key: 'lb_type', label: 'LB Type' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'type', label: 'Type' },
  { key: 'connection_number', label: 'Connection Number' },
  { key: 'data_transfer', label: 'Data Transfer' },
  { key: 'mode', label: 'Mode' },
  { key: 'lb_type', label: 'LB Type' },
  { key: 'display_lb_type', label: 'Display LB Type' },
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
    export-filename="greennode-lb-packages.csv"
    title="LB Packages"
  />
</template>
