<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/network-peerings'

const stateBadge = badgeColors({
  ACTIVE: badge.emerald,
  INACTIVE: badge.zinc,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'state', badge: stateBadge },
  { key: 'local_network', label: 'Local Network' },
  { key: 'network', label: 'Peer Network' },
  { key: 'exchange_subnet_routes', label: 'Subnet Routes', format: 'bool' },
  { key: 'export_custom_routes', label: 'Export Routes', format: 'bool' },
  { key: 'import_custom_routes', label: 'Import Routes', format: 'bool' },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'state' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'state', label: 'State' },
  { key: 'state_details', label: 'State Details' },
  { key: 'local_network', label: 'Local Network' },
  { key: 'network', label: 'Peer Network' },
  { key: 'stack_type', label: 'Stack Type' },
  { key: 'peer_mtu', label: 'Peer MTU' },
  { key: 'exchange_subnet_routes', label: 'Exchange Subnet Routes', format: 'bool' },
  { key: 'export_custom_routes', label: 'Export Custom Routes', format: 'bool' },
  { key: 'import_custom_routes', label: 'Import Custom Routes', format: 'bool' },
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
    search-placeholder="Search by name..."
    export-filename="gcp-compute-network-peerings.csv"
    drawer-title-key="name"
    title="Network Peerings"
  />
</template>
