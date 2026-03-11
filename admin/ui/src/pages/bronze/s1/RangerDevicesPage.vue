<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/ranger-devices'

const managedStateBadge = badgeColors({
  MANAGED: badge.emerald,
  UNMANAGED: badge.amber,
})

const columns: ColumnDef[] = [
  { key: 'local_ip', label: 'Local IP', format: 'mono' },
  { key: 'external_ip', label: 'External IP', format: 'mono' },
  { key: 'mac_address', label: 'MAC', format: 'mono' },
  { key: 'os_type', label: 'OS Type', badge: () => badge.zinc },
  { key: 'os_name', label: 'OS' },
  { key: 'device_type', label: 'Device Type', badge: () => badge.zinc },
  { key: 'device_function', label: 'Function' },
  { key: 'manufacturer' },
  { key: 'managed_state', label: 'Managed', badge: managedStateBadge },
  { key: 'site_name', label: 'Site' },
  { key: 'first_seen', label: 'First Seen', format: 'date' },
  { key: 'last_seen', label: 'Last Seen', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'os_type', label: 'OS Type' },
  { key: 'device_type', label: 'Device Type' },
  { key: 'managed_state', label: 'Managed' },
  { key: 'site_name', label: 'Site' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'local_ip', label: 'Local IP', mono: true },
  { key: 'external_ip', label: 'External IP', mono: true },
  { key: 'mac_address', label: 'MAC Address', mono: true },
  { key: 'os_type', label: 'OS Type' },
  { key: 'os_name', label: 'OS Name' },
  { key: 'os_version', label: 'OS Version' },
  { key: 'device_type', label: 'Device Type' },
  { key: 'device_function', label: 'Device Function' },
  { key: 'manufacturer', label: 'Manufacturer' },
  { key: 'managed_state', label: 'Managed State' },
  { key: 'agent_id', label: 'Agent ID', mono: true },
  { key: 'domain', label: 'Domain' },
  { key: 'site_name', label: 'Site' },
  { key: 'device_review', label: 'Device Review' },
  { key: 'subnet_address', label: 'Subnet Address' },
  { key: 'gateway_ip_address', label: 'Gateway IP', mono: true },
  { key: 'gateway_mac_address', label: 'Gateway MAC', mono: true },
  { key: 'network_name', label: 'Network Name' },
  { key: 'has_identity', label: 'Has Identity', format: 'bool' },
  { key: 'fingerprint_score', label: 'Fingerprint Score', transform: (v) => v != null ? String(v) : '' },
  { key: 'first_seen', label: 'First Seen', format: 'date' },
  { key: 'last_seen', label: 'Last Seen', format: 'date' },
  { key: 'first_collected_at', label: 'First Collected', format: 'date' },
  { key: 'collected_at', label: 'Last Collected', format: 'date' },
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
    export-filename="s1-ranger-devices.csv"
    drawer-title-key="local_ip"
    title="Ranger Devices"
  />
</template>
