<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/network/peerings'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
  ERROR: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'from_vpc_id', label: 'From VPC', format: 'mono' },
  { key: 'from_cidr', label: 'From CIDR', format: 'mono' },
  { key: 'end_vpc_id', label: 'End VPC', format: 'mono' },
  { key: 'end_cidr', label: 'End CIDR', format: 'mono' },
  { key: 'region' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'from_vpc_id', label: 'From VPC ID', mono: true },
  { key: 'from_cidr', label: 'From CIDR', mono: true },
  { key: 'end_vpc_id', label: 'End VPC ID', mono: true },
  { key: 'end_cidr', label: 'End CIDR', mono: true },
  { key: 'region', label: 'Region' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'created_at', label: 'Created', format: 'date' },
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
    export-filename="greennode-peerings.csv"
    title="Peerings"
  />
</template>
