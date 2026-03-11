<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/ranger-gateways'

const columns: ColumnDef[] = [
  { key: 'ip', label: 'IP', format: 'mono' },
  { key: 'mac_address', label: 'MAC', format: 'mono' },
  { key: 'external_ip', label: 'External IP', format: 'mono' },
  { key: 'manufacturer' },
  { key: 'network_name', label: 'Network' },
  { key: 'account_name', label: 'Account' },
  { key: 'number_of_agents', label: 'Agents', format: 'number' },
  { key: 'number_of_rangers', label: 'Rangers', format: 'number' },
  { key: 'connected_rangers', label: 'Connected', format: 'number' },
  { key: 'allow_scan', label: 'Scan', badge: (v) => v ? badge.emerald : badge.amber, transform: (v) => v ? 'Yes' : 'No' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'network_name', label: 'Network' },
  { key: 'account_name', label: 'Account' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'ip', label: 'IP', mono: true },
  { key: 'mac_address', label: 'MAC Address', mono: true },
  { key: 'external_ip', label: 'External IP', mono: true },
  { key: 'manufacturer', label: 'Manufacturer' },
  { key: 'network_name', label: 'Network' },
  { key: 'account_name', label: 'Account' },
  { key: 'account_id', label: 'Account ID', mono: true },
  { key: 'site_id', label: 'Site ID', mono: true },
  { key: 'number_of_agents', label: 'Agents', transform: (v) => v != null ? String(v) : '' },
  { key: 'number_of_rangers', label: 'Rangers', transform: (v) => v != null ? String(v) : '' },
  { key: 'connected_rangers', label: 'Connected Rangers', transform: (v) => v != null ? String(v) : '' },
  { key: 'total_agents', label: 'Total Agents', transform: (v) => v != null ? String(v) : '' },
  { key: 'agent_percentage', label: 'Agent Percentage', transform: (v) => v != null ? String(v) : '' },
  { key: 'allow_scan', label: 'Allow Scan', format: 'bool' },
  { key: 'archived', label: 'Archived', format: 'bool' },
  { key: 'new_network', label: 'New Network', format: 'bool' },
  { key: 'inherit_settings', label: 'Inherit Settings', format: 'bool' },
  { key: 'tcp_port_scan', label: 'TCP Port Scan', format: 'bool' },
  { key: 'udp_port_scan', label: 'UDP Port Scan', format: 'bool' },
  { key: 'icmp_scan', label: 'ICMP Scan', format: 'bool' },
  { key: 'smb_scan', label: 'SMB Scan', format: 'bool' },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'expiry_date', label: 'Expiry Date', format: 'date' },
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
    search-key="q"
    search-placeholder="Search by IP..."
    export-filename="s1-ranger-gateways.csv"
    drawer-title-key="ip"
    title="Ranger Gateways"
  />
</template>
