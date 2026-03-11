<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/backend-service-backends'

const columns: ColumnDef[] = [
  { key: 'service_name', label: 'Service', format: 'bold' },
  { key: 'group', label: 'Backend Group', transform: shortPath },
  { key: 'balancing_mode', label: 'Balancing Mode', format: 'mono' },
  { key: 'load_balancing_scheme', label: 'Scheme' },
  { key: 'protocol', format: 'mono' },
  { key: 'failover', format: 'bool' },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'balancing_mode', label: 'Balancing Mode' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'service_name', label: 'Service' },
  { key: 'group', label: 'Backend Group', mono: true },
  { key: 'balancing_mode', label: 'Balancing Mode' },
  { key: 'capacity_scaler', label: 'Capacity Scaler' },
  { key: 'failover', label: 'Failover', format: 'bool' },
  { key: 'max_rate', label: 'Max Rate' },
  { key: 'max_utilization', label: 'Max Utilization' },
  { key: 'load_balancing_scheme', label: 'LB Scheme' },
  { key: 'protocol', label: 'Protocol' },
  { key: 'project_id', label: 'Project ID', mono: true },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-id"
    search-key="service_name"
    search-placeholder="Search by service name..."
    export-filename="gcp-compute-backend-service-backends.csv"
    drawer-title-key="service_name"
    title="Backend Service Backends"
  />
</template>
