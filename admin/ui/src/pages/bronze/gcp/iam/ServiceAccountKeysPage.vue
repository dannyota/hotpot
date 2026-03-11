<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/iam/service-account-keys'

const keyTypeBadge = badgeColors({
  USER_MANAGED: badge.blue,
  SYSTEM_MANAGED: badge.zinc,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'mono' },
  { key: 'service_account_email', label: 'Service Account' },
  { key: 'key_origin', label: 'Origin', badge: () => badge.zinc },
  { key: 'key_type', label: 'Type', badge: keyTypeBadge },
  { key: 'key_algorithm', label: 'Algorithm' },
  { key: 'valid_after_time', label: 'Valid After', format: 'date' },
  { key: 'valid_before_time', label: 'Valid Before', format: 'date' },
  { key: 'disabled', badge: (v) => v ? badge.amber : badge.emerald, transform: (v) => v ? 'Yes' : 'No' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'key_type', label: 'Type' },
  { key: 'key_origin', label: 'Origin' },
  { key: 'disabled' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name', mono: true },
  { key: 'service_account_email', label: 'Service Account' },
  { key: 'key_origin', label: 'Origin' },
  { key: 'key_type', label: 'Type' },
  { key: 'key_algorithm', label: 'Algorithm' },
  { key: 'valid_after_time', label: 'Valid After', format: 'date' },
  { key: 'valid_before_time', label: 'Valid Before', format: 'date' },
  { key: 'disabled', label: 'Disabled', format: 'bool' },
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
    export-filename="gcp-iam-service-account-keys.csv"
    drawer-title-key="name"
    title="Service Account Keys"
  />
</template>
