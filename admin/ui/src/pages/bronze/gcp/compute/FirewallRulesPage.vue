<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/firewall-rules'

const directionBadge = badgeColors({
  INGRESS: badge.blue,
  EGRESS: badge.purple,
})

const columns: ColumnDef[] = [
  { key: 'firewall_name', label: 'Firewall', format: 'bold' },
  { key: 'ip_protocol', label: 'Protocol', format: 'mono' },
  { key: 'direction', badge: directionBadge },
  { key: 'priority', format: 'number' },
  { key: 'network', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'ip_protocol', label: 'Protocol' },
  { key: 'direction' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'firewall_name', label: 'Firewall' },
  { key: 'ip_protocol', label: 'Protocol' },
  { key: 'ports_json', label: 'Ports' },
  { key: 'direction', label: 'Direction' },
  { key: 'priority', label: 'Priority' },
  { key: 'network', label: 'Network', transform: shortPath },
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
    search-key="firewall_name"
    search-placeholder="Search by firewall name..."
    export-filename="gcp-compute-firewall-allow-rules.csv"
    drawer-title-key="firewall_name"
    title="Firewall Allow Rules"
  />
</template>
