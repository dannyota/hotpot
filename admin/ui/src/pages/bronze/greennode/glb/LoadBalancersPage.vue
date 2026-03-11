<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/glb/load-balancers'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
  ERROR: badge.red,
})

const typeBadge = badgeColors({
  HTTP: badge.blue,
  TCP: badge.amber,
  UDP: badge.purple,
}, badge.zinc)

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'package', label: 'Package' },
  { key: 'type', badge: typeBadge },
  { key: 'description' },
  { key: 'user_id', label: 'User ID', format: 'mono' },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'updated_at_api', label: 'Updated', format: 'date' },
  { key: 'deleted_at_api', label: 'Deleted', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'type' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'package', label: 'Package' },
  { key: 'type', label: 'Type' },
  { key: 'description', label: 'Description' },
  { key: 'user_id', label: 'User ID', mono: true },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'updated_at_api', label: 'Updated', format: 'date' },
  { key: 'deleted_at_api', label: 'Deleted', format: 'date' },
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
    export-filename="greennode-glb-load-balancers.csv"
    title="Global Load Balancers"
  />
</template>
