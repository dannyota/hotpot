<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/silver/inventory/api-endpoints'

const accessLevelBadge = badgeColors({
  PUBLIC: badge.emerald,
  PROTECTED: badge.amber,
  PRIVATE: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'uri_pattern', label: 'URI Pattern', format: 'mono' },
  { key: 'name', format: 'bold' },
  { key: 'service', badge: () => badge.zinc },
  { key: 'access_level', label: 'Access Level', badge: accessLevelBadge },
  { key: 'is_active', label: 'Active', badge: (v) => v ? badge.emerald : badge.amber, transform: (v) => v ? 'Yes' : 'No' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
  { key: 'normalized_at', format: 'date' },
]

const filters: FilterDef[] = [
  { key: 'service' },
  { key: 'access_level', label: 'Access Level' },
  { key: 'is_active', label: 'Active', bool: true },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'service', label: 'Service' },
  { key: 'uri_pattern', label: 'URI Pattern', mono: true },
  { key: 'is_active', label: 'Active', format: 'bool' },
  { key: 'access_level', label: 'Access Level' },
  { key: 'collected_at', label: 'Last Seen', format: 'date' },
  { key: 'first_collected_at', label: 'First Seen', format: 'date' },
  { key: 'normalized_at', label: 'Normalized', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-collected_at"
    search-key="uri_pattern"
    search-placeholder="Search by URI pattern..."
    export-filename="silver-api-endpoints.csv"
    drawer-title-key="uri_pattern"
    title="API Endpoints"
  />
</template>
