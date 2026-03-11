<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/negs'

const typeBadge = badgeColors({
  GCE_VM_IP_PORT: badge.blue,
  GCE_VM_IP: badge.blue,
  INTERNET_IP_PORT: badge.purple,
  INTERNET_FQDN_PORT: badge.purple,
  SERVERLESS: badge.emerald,
  NON_GCP_PRIVATE_IP_PORT: badge.amber,
  PRIVATE_SERVICE_CONNECT: badge.purple,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'network_endpoint_type', label: 'Type', badge: typeBadge },
  { key: 'zone', transform: shortPath, format: 'mono' },
  { key: 'default_port', label: 'Port', format: 'mono' },
  { key: 'size', format: 'number' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'network_endpoint_type', label: 'Type' },
  { key: 'zone' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'network_endpoint_type', label: 'Type' },
  { key: 'zone', label: 'Zone', transform: shortPath },
  { key: 'default_port', label: 'Default Port' },
  { key: 'size', label: 'Size' },
  { key: 'region', label: 'Region', transform: shortPath },
  { key: 'network', label: 'Network', transform: shortPath },
  { key: 'subnetwork', label: 'Subnetwork', transform: shortPath },
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
    export-filename="gcp-compute-negs.csv"
    title="Network Endpoint Groups"
  />
</template>
