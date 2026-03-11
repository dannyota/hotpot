<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/iam/service-accounts'

const columns: ColumnDef[] = [
  { key: 'email', format: 'mono', maxWidth: 320 },
  { key: 'display_name', label: 'Display Name' },
  { key: 'disabled', badge: (v) => v ? badge.red : badge.emerald, transform: (v) => v ? 'Disabled' : 'Active' },
  { key: 'description' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'disabled' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name', mono: true },
  { key: 'email', label: 'Email', mono: true },
  { key: 'display_name', label: 'Display Name' },
  { key: 'disabled', label: 'Disabled', format: 'bool' },
  { key: 'description', label: 'Description' },
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
    export-filename="gcp-iam-service-accounts.csv"
    drawer-title-key="email"
    title="IAM Service Accounts"
  />
</template>
