<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { shortPath, shortEmail } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/instance-service-accounts'

const columns: ColumnDef[] = [
  { key: 'instance_name', label: 'Instance', format: 'bold' },
  { key: 'email', transform: shortEmail, format: 'mono', maxWidth: 320 },
  { key: 'instance_zone', label: 'Zone', transform: shortPath, format: 'mono' },
  { key: 'project_id', label: 'Project', format: 'mono', maxWidth: 240 },
]

const filters: FilterDef[] = [
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'email', label: 'Email', mono: true },
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
    search-key="email"
    search-placeholder="Search by email..."
    export-filename="gcp-compute-instance-service-accounts.csv"
    drawer-title-key="email"
    title="Instance Service Accounts"
  />
</template>
