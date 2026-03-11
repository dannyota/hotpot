<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/storage/buckets'

const storageClassBadge = badgeColors({
  STANDARD: badge.emerald,
  NEARLINE: badge.blue,
  ARCHIVE: badge.purple,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'location' },
  { key: 'storage_class', label: 'Storage Class', badge: storageClassBadge },
  { key: 'default_event_based_hold', label: 'Event Hold', badge: (v) => v ? badge.amber : badge.zinc, transform: (v) => v ? 'Yes' : 'No' },
  { key: 'time_created', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'location' },
  { key: 'storage_class', label: 'Class' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'location', label: 'Location' },
  { key: 'storage_class', label: 'Storage Class' },
  { key: 'default_event_based_hold', label: 'Event-Based Hold', format: 'bool' },
  { key: 'time_created', label: 'Created', format: 'date' },
  { key: 'project_id', label: 'Project ID', mono: true },
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
    search-placeholder="Search by name..."
    export-filename="gcp-storage-buckets.csv"
    drawer-title-key="name"
    title="Storage Buckets"
  />
</template>
