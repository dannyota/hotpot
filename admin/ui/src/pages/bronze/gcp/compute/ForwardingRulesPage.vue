<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/forwarding-rules'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'ip_address', label: 'IP Address', format: 'mono' },
  { key: 'ip_protocol', label: 'Protocol' },
  { key: 'port_range', label: 'Port Range', format: 'mono' },
  { key: 'load_balancing_scheme', label: 'LB Scheme' },
  { key: 'network_tier', label: 'Tier' },
  { key: 'region', transform: shortPath },
  { key: 'target', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'ip_protocol', label: 'Protocol' },
  { key: 'load_balancing_scheme', label: 'LB Scheme' },
  { key: 'network_tier', label: 'Tier' },
  { key: 'region' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'ip_address', label: 'IP Address', mono: true },
  { key: 'ip_protocol', label: 'Protocol' },
  { key: 'port_range', label: 'Port Range' },
  { key: 'load_balancing_scheme', label: 'LB Scheme' },
  { key: 'network_tier', label: 'Tier' },
  { key: 'region', label: 'Region', transform: shortPath },
  { key: 'target', label: 'Target', transform: shortPath },
  { key: 'network', label: 'Network', transform: shortPath },
  { key: 'subnetwork', label: 'Subnetwork', transform: shortPath },
  { key: 'service_name', label: 'Service Name' },
  { key: 'service_label', label: 'Service Label' },
  { key: 'all_ports', label: 'All Ports', format: 'bool' },
  { key: 'allow_global_access', label: 'Allow Global Access', format: 'bool' },
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
    export-filename="gcp-compute-forwarding-rules.csv"
    title="Forwarding Rules"
  />
</template>
