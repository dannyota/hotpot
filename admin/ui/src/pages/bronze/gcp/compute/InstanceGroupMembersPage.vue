<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/instance-group-members'

const statusBadge = badgeColors({
  RUNNING: badge.emerald,
  STOPPED: badge.zinc,
  TERMINATED: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'instance_name', label: 'Instance', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'group_name', label: 'Group' },
  { key: 'group_zone', label: 'Zone', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'instance_name', label: 'Instance' },
  { key: 'status', label: 'Status' },
  { key: 'group_name', label: 'Group' },
  { key: 'group_zone', label: 'Zone', transform: shortPath },
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
    search-key="instance_name"
    search-placeholder="Search by instance name..."
    export-filename="gcp-compute-instance-group-members.csv"
    drawer-title-key="instance_name"
    title="Instance Group Members"
  />
</template>
