<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors, formatUnit } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/instance-disks'

const columns: ColumnDef[] = [
  { key: 'device_name', label: 'Device', format: 'bold' },
  { key: 'boot', format: 'bool' },
  { key: 'mode', format: 'mono' },
  { key: 'type', format: 'mono' },
  { key: 'disk_size_gb', label: 'Size', format: 'number', transform: (v) => formatUnit(v, 'GB') },
  { key: 'auto_delete', label: 'Auto Delete', format: 'bool' },
  { key: 'instance_name', label: 'Instance' },
  { key: 'instance_zone', label: 'Zone', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'boot' },
  { key: 'mode' },
  { key: 'type' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'device_name', label: 'Device Name' },
  { key: 'source', label: 'Source', mono: true, transform: shortPath },
  { key: 'index', label: 'Index' },
  { key: 'boot', label: 'Boot', format: 'bool' },
  { key: 'auto_delete', label: 'Auto Delete', format: 'bool' },
  { key: 'mode', label: 'Mode' },
  { key: 'type', label: 'Type' },
  { key: 'disk_size_gb', label: 'Size', transform: (v) => formatUnit(v, 'GB') },
  { key: 'instance_name', label: 'Instance' },
  { key: 'instance_zone', label: 'Zone', transform: shortPath },
  { key: 'project_id', label: 'Project ID', mono: true },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-id"
    search-key="device_name"
    search-placeholder="Search by device name..."
    export-filename="gcp-compute-instance-disks.csv"
    drawer-title-key="device_name"
    title="Instance Disks"
  />
</template>
