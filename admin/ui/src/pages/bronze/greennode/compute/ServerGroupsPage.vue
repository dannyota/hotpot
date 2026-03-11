<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/compute/server-groups'

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'description' },
  { key: 'policy_id', label: 'Policy ID', format: 'mono' },
  { key: 'policy_name', label: 'Policy Name' },
  { key: 'region' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'policy_id', label: 'Policy ID', mono: true },
  { key: 'policy_name', label: 'Policy Name' },
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
    export-filename="greennode-server-groups.csv"
    title="Server Groups"
  />
</template>
