<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors, humanize } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/addresses'

const statusBadge = badgeColors({
  IN_USE: badge.emerald,
  RESERVED: badge.blue,
  RESERVING: badge.amber,
})

const typeBadge = badgeColors({
  EXTERNAL: badge.purple,
  INTERNAL: badge.blue,
})

const purposeBadge = badgeColors({
  GCE_ENDPOINT: badge.blue,
  PRIVATE_SERVICE_CONNECT: badge.purple,
  NAT_AUTO: badge.amber,
  VPC_PEERING: badge.emerald,
  SHARED_LOADBALANCER_VIP: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'address', format: 'mono' },
  { key: 'status', badge: statusBadge },
  { key: 'address_type', label: 'Type', badge: typeBadge },
  { key: 'ip_version', label: 'IP Version', format: 'mono' },
  { key: 'region', transform: shortPath },
  { key: 'network_tier', label: 'Tier', format: 'mono' },
  { key: 'purpose', badge: purposeBadge },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'address_type', label: 'Type' },
  { key: 'region' },
  { key: 'network_tier', label: 'Tier' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'address', label: 'Address', mono: true },
  { key: 'status', label: 'Status' },
  { key: 'address_type', label: 'Address Type' },
  { key: 'ip_version', label: 'IP Version' },
  { key: 'region', label: 'Region', transform: shortPath },
  { key: 'network_tier', label: 'Network Tier' },
  { key: 'purpose', label: 'Purpose', transform: humanize },
  { key: 'network', label: 'Network', transform: shortPath },
  { key: 'subnetwork', label: 'Subnetwork', transform: shortPath },
  { key: 'prefix_length', label: 'Prefix Length' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'self_link', label: 'Self Link', mono: true },
  { key: 'creation_timestamp', label: 'Created' },
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
    default-sort="-creation_timestamp"
    export-filename="gcp-compute-addresses.csv"
    title="Addresses"
  />
</template>
