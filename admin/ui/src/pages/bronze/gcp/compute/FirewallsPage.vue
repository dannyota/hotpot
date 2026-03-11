<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/firewalls'

const directionBadge = badgeColors({
  INGRESS: badge.blue,
  EGRESS: badge.purple,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'direction', badge: directionBadge },
  { key: 'priority', format: 'number' },
  { key: 'disabled', format: 'bool' },
  { key: 'network', transform: shortPath },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'direction' },
  { key: 'network' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'direction', label: 'Direction' },
  { key: 'priority', label: 'Priority' },
  { key: 'disabled', label: 'Disabled', format: 'bool' },
  { key: 'network', label: 'Network', transform: shortPath },
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
    export-filename="gcp-compute-firewalls.csv"
    title="Firewalls"
  />
</template>
