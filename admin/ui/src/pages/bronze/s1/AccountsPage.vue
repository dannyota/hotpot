<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/accounts'

const stateBadge = badgeColors({
  ACTIVE: badge.emerald,
  EXPIRED: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'state', badge: stateBadge },
  { key: 'account_type', label: 'Type', badge: () => badge.zinc },
  { key: 'active_agents', label: 'Active Agents', format: 'number' },
  { key: 'total_licenses', label: 'Licenses', format: 'number' },
  { key: 'usage_type', label: 'Usage' },
  { key: 'billing_mode', label: 'Billing' },
  { key: 'unlimited_expiration', label: 'Unlimited', format: 'bool' },
  { key: 'api_created_at', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'state' },
  { key: 'account_type', label: 'Type' },
  { key: 'usage_type', label: 'Usage' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'state', label: 'State' },
  { key: 'account_type', label: 'Type' },
  { key: 'active_agents', label: 'Active Agents', transform: (v) => v != null ? String(v) : '' },
  { key: 'total_licenses', label: 'Total Licenses', transform: (v) => v != null ? String(v) : '' },
  { key: 'usage_type', label: 'Usage' },
  { key: 'billing_mode', label: 'Billing' },
  { key: 'unlimited_expiration', label: 'Unlimited', format: 'bool' },
  { key: 'creator', label: 'Creator' },
  { key: 'expiration', label: 'Expiration', format: 'date' },
  { key: 'api_created_at', label: 'Created', format: 'date' },
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
    export-filename="s1-accounts.csv"
    drawer-title-key="name"
    title="Accounts"
  />
</template>
