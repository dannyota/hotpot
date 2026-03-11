<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/portal/quotas'

const typeBadge = badgeColors({}, badge.blue)

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'description' },
  { key: 'type', badge: typeBadge },
  { key: 'limit_value', label: 'Limit', format: 'number' },
  { key: 'used_value', label: 'Used', format: 'number' },
  { key: 'region' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'type' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'type', label: 'Type' },
  { key: 'limit_value', label: 'Limit' },
  { key: 'used_value', label: 'Used' },
  { key: 'region', label: 'Region' },
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
    export-filename="greennode-portal-quotas.csv"
    title="Quotas"
  />
</template>
