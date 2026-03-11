<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/silver/inventory/machines'

const osTypeBadge = badgeColors({
  WINDOWS: badge.blue,
  LINUX: badge.amber,
  MACOS: badge.purple,
})

const statusBadge = badgeColors({
  RUNNING: badge.emerald,
  STOPPED: badge.zinc,
  ERROR: badge.red,
})

const envBadge = badgeColors({
  PRODUCTION: badge.red,
  UAT: badge.amber,
})

const columns: ColumnDef[] = [
  { key: 'hostname', format: 'bold' },
  { key: 'os_type', label: 'OS Type', badge: osTypeBadge },
  { key: 'os_name', label: 'OS Name' },
  { key: 'status', badge: statusBadge },
  { key: 'environment', badge: envBadge },
  { key: 'cloud_project', label: 'Project' },
  { key: 'cloud_zone', label: 'Zone' },
  { key: 'internal_ip', label: 'Internal IP', format: 'mono' },
  { key: 'external_ip', label: 'External IP', format: 'mono' },
  { key: 'created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'os_type', label: 'OS Type' },
  { key: 'status' },
  { key: 'environment' },
  { key: 'cloud_project', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'hostname', label: 'Hostname' },
  { key: 'os_type', label: 'OS Type' },
  { key: 'os_name', label: 'OS Name' },
  { key: 'status', label: 'Status' },
  { key: 'environment', label: 'Environment' },
  { key: 'cloud_project', label: 'Project' },
  { key: 'cloud_zone', label: 'Zone' },
  { key: 'internal_ip', label: 'Internal IP', mono: true },
  { key: 'external_ip', label: 'External IP', mono: true },
  { key: 'created', label: 'Created', format: 'date' },
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
    default-sort="-created"
    search-key="hostname"
    search-placeholder="Search by hostname..."
    export-filename="silver-machines.csv"
    drawer-title-key="hostname"
    title="Machines"
  />
</template>
