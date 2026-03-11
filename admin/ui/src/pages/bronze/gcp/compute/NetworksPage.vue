<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/networks'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'routing_mode', label: 'Routing Mode' },
  { key: 'auto_create_subnetworks', label: 'Auto Subnets', format: 'bool' },
  { key: 'mtu', label: 'MTU', format: 'number' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'routing_mode', label: 'Routing' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'routing_mode', label: 'Routing Mode' },
  { key: 'auto_create_subnetworks', label: 'Auto Create Subnetworks', format: 'bool' },
  { key: 'mtu', label: 'MTU' },
  { key: 'gateway_ipv4', label: 'Gateway IPv4' },
  { key: 'enable_ula_internal_ipv6', label: 'ULA Internal IPv6', format: 'bool' },
  { key: 'internal_ipv6_range', label: 'Internal IPv6 Range' },
  { key: 'network_firewall_policy_enforcement_order', label: 'Firewall Policy Order' },
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
    export-filename="gcp-compute-networks.csv"
    title="Networks"
  />
</template>
