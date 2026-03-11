<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/groups'

const typeBadge = badgeColors({
  STATIC: badge.blue,
  DYNAMIC: badge.purple,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'type', badge: typeBadge },
  { key: 'site_id', label: 'Site ID', format: 'mono' },
  { key: 'is_default', label: 'Default', format: 'bool' },
  { key: 'inherits', format: 'bool' },
  { key: 'total_agents', label: 'Agents', format: 'number' },
  { key: 'rank', format: 'number' },
  { key: 'creator' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'type' },
  { key: 'is_default', label: 'Default' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'type', label: 'Type' },
  { key: 'site_id', label: 'Site ID', mono: true },
  { key: 'is_default', label: 'Default', format: 'bool' },
  { key: 'inherits', label: 'Inherits', format: 'bool' },
  { key: 'total_agents', label: 'Total Agents', transform: (v) => v != null ? String(v) : '' },
  { key: 'rank', label: 'Rank', transform: (v) => v != null ? String(v) : '' },
  { key: 'creator', label: 'Creator' },
  { key: 'creator_id', label: 'Creator ID', mono: true },
  { key: 'filter_name', label: 'Filter Name' },
  { key: 'filter_id', label: 'Filter ID', mono: true },
  { key: 'registration_token', label: 'Registration Token', mono: true },
  { key: 'api_created_at', label: 'API Created', format: 'date' },
  { key: 'api_updated_at', label: 'API Updated', format: 'date' },
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
    export-filename="s1-groups.csv"
    drawer-title-key="name"
    title="Groups"
  />
</template>
