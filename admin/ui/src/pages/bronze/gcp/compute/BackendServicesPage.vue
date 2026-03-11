<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/backend-services'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'load_balancing_scheme', label: 'LB Scheme' },
  { key: 'protocol' },
  { key: 'port', format: 'mono' },
  { key: 'region', transform: shortPath },
  { key: 'enable_cdn', label: 'CDN', format: 'bool' },
  { key: 'session_affinity', label: 'Session Affinity' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'load_balancing_scheme', label: 'LB Scheme' },
  { key: 'protocol' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'load_balancing_scheme', label: 'LB Scheme' },
  { key: 'protocol', label: 'Protocol' },
  { key: 'port', label: 'Port' },
  { key: 'port_name', label: 'Port Name' },
  { key: 'timeout_sec', label: 'Timeout (sec)' },
  { key: 'region', label: 'Region', transform: shortPath },
  { key: 'network', label: 'Network', transform: shortPath },
  { key: 'enable_cdn', label: 'CDN', format: 'bool' },
  { key: 'session_affinity', label: 'Session Affinity' },
  { key: 'locality_lb_policy', label: 'Locality LB Policy' },
  { key: 'compression_mode', label: 'Compression Mode' },
  { key: 'security_policy', label: 'Security Policy', transform: shortPath },
  { key: 'edge_security_policy', label: 'Edge Security Policy', transform: shortPath },
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
    export-filename="gcp-compute-backend-services.csv"
    title="Backend Services"
  />
</template>
