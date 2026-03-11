<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/sites'

const stateBadge = badgeColors({
  ACTIVE: badge.emerald,
  EXPIRED: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'state', badge: stateBadge },
  { key: 'site_type', label: 'Type', badge: () => badge.zinc },
  { key: 'account_name', label: 'Account' },
  { key: 'active_licenses', label: 'Active Licenses', format: 'number' },
  { key: 'total_licenses', label: 'Total Licenses', format: 'number' },
  { key: 'health_status', label: 'Health', badge: (v) => v ? badge.emerald : badge.red, transform: (v) => v ? 'Healthy' : 'Unhealthy' },
  { key: 'is_default', label: 'Default', format: 'bool' },
  { key: 'suite' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'state' },
  { key: 'site_type', label: 'Type' },
  { key: 'account_name', label: 'Account' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'state', label: 'State' },
  { key: 'site_type', label: 'Type' },
  { key: 'account_name', label: 'Account' },
  { key: 'account_id', label: 'Account ID' },
  { key: 'active_licenses', label: 'Active Licenses', transform: (v) => v != null ? String(v) : '' },
  { key: 'total_licenses', label: 'Total Licenses', transform: (v) => v != null ? String(v) : '' },
  { key: 'unlimited_licenses', label: 'Unlimited Licenses', format: 'bool' },
  { key: 'health_status', label: 'Health Status', transform: (v) => v ? 'Healthy' : 'Unhealthy' },
  { key: 'is_default', label: 'Default', format: 'bool' },
  { key: 'suite', label: 'Suite' },
  { key: 'creator', label: 'Creator' },
  { key: 'description', label: 'Description' },
  { key: 'sku', label: 'SKU' },
  { key: 'usage_type', label: 'Usage Type' },
  { key: 'expiration', label: 'Expiration', format: 'date' },
  { key: 'api_created_at', label: 'API Created At', format: 'date' },
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
    export-filename="s1-sites.csv"
    drawer-title-key="name"
    title="Sites"
  />
</template>
