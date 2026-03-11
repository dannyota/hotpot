<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/ranger-settings'

const columns: ColumnDef[] = [
  { key: 'account_id', label: 'Account ID', format: 'mono' },
  { key: 'scope_id', label: 'Scope ID', format: 'mono' },
  { key: 'enabled', format: 'bool' },
  { key: 'tcp_port_scan', label: 'TCP', format: 'bool' },
  { key: 'udp_port_scan', label: 'UDP', format: 'bool' },
  { key: 'icmp_scan', label: 'ICMP', format: 'bool' },
  { key: 'smb_scan', label: 'SMB', format: 'bool' },
  { key: 'mdns_scan', label: 'mDNS', format: 'bool' },
  { key: 'rdns_scan', label: 'rDNS', format: 'bool' },
  { key: 'snmp_scan', label: 'SNMP', format: 'bool' },
  { key: 'use_periodic_snapshots', label: 'Snapshots', format: 'bool' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'enabled' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'account_id', label: 'Account ID', mono: true },
  { key: 'scope_id', label: 'Scope ID', mono: true },
  { key: 'enabled', label: 'Enabled', format: 'bool' },
  { key: 'use_periodic_snapshots', label: 'Use Periodic Snapshots', format: 'bool' },
  { key: 'snapshot_period', label: 'Snapshot Period', transform: (v) => v != null ? String(v) : '' },
  { key: 'network_decommission_value', label: 'Network Decommission Value', transform: (v) => v != null ? String(v) : '' },
  { key: 'min_agents_in_network_to_scan', label: 'Min Agents in Network to Scan', transform: (v) => v != null ? String(v) : '' },
  { key: 'tcp_port_scan', label: 'TCP Port Scan', format: 'bool' },
  { key: 'udp_port_scan', label: 'UDP Port Scan', format: 'bool' },
  { key: 'icmp_scan', label: 'ICMP Scan', format: 'bool' },
  { key: 'smb_scan', label: 'SMB Scan', format: 'bool' },
  { key: 'mdns_scan', label: 'mDNS Scan', format: 'bool' },
  { key: 'rdns_scan', label: 'rDNS Scan', format: 'bool' },
  { key: 'snmp_scan', label: 'SNMP Scan', format: 'bool' },
  { key: 'multi_scan_ssdp', label: 'Multi Scan SSDP', format: 'bool' },
  { key: 'use_full_dns_scan', label: 'Use Full DNS Scan', format: 'bool' },
  { key: 'scan_only_local_subnets', label: 'Scan Only Local Subnets', format: 'bool' },
  { key: 'auto_enable_networks', label: 'Auto Enable Networks', format: 'bool' },
  { key: 'combine_devices', label: 'Combine Devices', format: 'bool' },
  { key: 'new_network_in_hours', label: 'New Network in Hours', transform: (v) => v != null ? String(v) : '' },
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
    search-key="account_id"
    search-placeholder="Search by account..."
    export-filename="s1-ranger-settings.csv"
    drawer-title-key="account_id"
    title="Ranger Settings"
  />
</template>
