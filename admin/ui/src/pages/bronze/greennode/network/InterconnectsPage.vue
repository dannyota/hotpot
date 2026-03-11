<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/network/interconnects'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
  ERROR: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'type_name', label: 'Type' },
  { key: 'circuit_id', label: 'Circuit ID', format: 'mono' },
  { key: 'gw_vip', label: 'GW VIP', format: 'mono' },
  { key: 'region' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'type_name', label: 'Type' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'status', label: 'Status' },
  { key: 'type_name', label: 'Type' },
  { key: 'type_id', label: 'Type ID', mono: true },
  { key: 'circuit_id', label: 'Circuit ID', mono: true },
  { key: 'enable_gw2', label: 'Enable GW2' },
  { key: 'gw01_ip', label: 'GW01 IP', mono: true },
  { key: 'gw02_ip', label: 'GW02 IP', mono: true },
  { key: 'gw_vip', label: 'GW VIP', mono: true },
  { key: 'remote_gw01_ip', label: 'Remote GW01 IP', mono: true },
  { key: 'remote_gw02_ip', label: 'Remote GW02 IP', mono: true },
  { key: 'package_id', label: 'Package ID', mono: true },
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
    export-filename="greennode-interconnects.csv"
    title="Interconnects"
  />
</template>
