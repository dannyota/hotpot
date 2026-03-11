<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/subnetwork-secondary-ranges'

const columns: ColumnDef[] = [
  { key: 'range_name', label: 'Range Name', format: 'bold' },
  { key: 'ip_cidr_range', label: 'CIDR', format: 'mono' },
  { key: 'subnetwork_name', label: 'Subnetwork' },
  { key: 'region', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'range_name', label: 'Range Name' },
  { key: 'ip_cidr_range', label: 'CIDR' },
  { key: 'subnetwork_name', label: 'Subnetwork' },
  { key: 'region', label: 'Region', transform: shortPath },
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
    search-key="range_name"
    search-placeholder="Search by range name..."
    export-filename="gcp-compute-subnetwork-secondary-ranges.csv"
    drawer-title-key="range_name"
    title="Subnetwork Secondary Ranges"
  />
</template>
