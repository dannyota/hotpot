<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/network-discoveries'

const assetStatusBadge = badgeColors({
  MANAGED: badge.emerald,
  UNMANAGED: badge.amber,
  BLOCKED: badge.red,
})

const infectionStatusBadge = badgeColors({
  CLEAN: badge.emerald,
  INFECTED: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'ip_address', label: 'IP', format: 'mono' },
  { key: 'category', badge: () => badge.zinc },
  { key: 'sub_category', label: 'Sub-Category' },
  { key: 'os' },
  { key: 'os_family', label: 'OS Family' },
  { key: 'manufacturer' },
  { key: 'asset_status', label: 'Status', badge: assetStatusBadge },
  { key: 'infection_status', label: 'Infection', badge: infectionStatusBadge },
  { key: 'device_review', label: 'Review' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'category' },
  { key: 'os_family', label: 'OS Family' },
  { key: 'asset_status', label: 'Status' },
  { key: 'infection_status', label: 'Infection' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'ip_address', label: 'IP Address', mono: true },
  { key: 'domain', label: 'Domain' },
  { key: 'serial_number', label: 'Serial Number', mono: true },
  { key: 'category', label: 'Category' },
  { key: 'sub_category', label: 'Sub-Category' },
  { key: 'resource_type', label: 'Resource Type' },
  { key: 'os', label: 'OS' },
  { key: 'os_family', label: 'OS Family' },
  { key: 'os_version', label: 'OS Version' },
  { key: 'os_name_version', label: 'OS Name & Version' },
  { key: 'architecture', label: 'Architecture' },
  { key: 'manufacturer', label: 'Manufacturer' },
  { key: 'cpu', label: 'CPU' },
  { key: 'memory_readable', label: 'Memory' },
  { key: 'network_name', label: 'Network Name' },
  { key: 'asset_status', label: 'Asset Status' },
  { key: 'asset_criticality', label: 'Asset Criticality' },
  { key: 'asset_environment', label: 'Asset Environment' },
  { key: 'infection_status', label: 'Infection Status' },
  { key: 'device_review', label: 'Device Review' },
  { key: 'detected_from_site', label: 'Detected From Site' },
  { key: 's1_account_name', label: 'S1 Account' },
  { key: 's1_site_name', label: 'S1 Site' },
  { key: 's1_group_name', label: 'S1 Group' },
  { key: 'first_seen_dt', label: 'First Seen (API)', format: 'date' },
  { key: 'last_update_dt', label: 'Last Updated', format: 'date' },
  { key: 'last_active_dt', label: 'Last Active', format: 'date' },
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
    search-placeholder="Search by name..."
    export-filename="s1-network-discoveries.csv"
    drawer-title-key="name"
    title="Network Discoveries"
  />
</template>
