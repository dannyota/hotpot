<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors, humanize } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/subnetworks'

const purposeBadge = badgeColors({
  PRIVATE: badge.blue,
  PRIVATE_SERVICE_CONNECT: badge.purple,
  REGIONAL_MANAGED_PROXY: badge.amber,
  GLOBAL_MANAGED_PROXY: badge.amber,
  INTERNAL_HTTPS_LOAD_BALANCER: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'region', transform: shortPath },
  { key: 'network', transform: shortPath },
  { key: 'ip_cidr_range', label: 'CIDR', format: 'mono' },
  { key: 'purpose', badge: purposeBadge },
  { key: 'stack_type', label: 'Stack', format: 'mono' },
  { key: 'private_ip_google_access', label: 'Private Google', format: 'bool' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'region' },
  { key: 'network' },
  { key: 'purpose' },
  { key: 'stack_type', label: 'Stack' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'region', label: 'Region', transform: shortPath },
  { key: 'network', label: 'Network', transform: shortPath },
  { key: 'ip_cidr_range', label: 'IP CIDR Range' },
  { key: 'gateway_address', label: 'Gateway Address' },
  { key: 'purpose', label: 'Purpose', transform: humanize },
  { key: 'role', label: 'Role' },
  { key: 'stack_type', label: 'Stack Type' },
  { key: 'ipv6_access_type', label: 'IPv6 Access Type' },
  { key: 'private_ip_google_access', label: 'Private Google Access', format: 'bool' },
  { key: 'private_ipv6_google_access', label: 'Private IPv6 Google Access' },
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
    export-filename="gcp-compute-subnetworks.csv"
    title="Subnetworks"
  />
</template>
