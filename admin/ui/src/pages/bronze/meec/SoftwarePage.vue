<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/meec/inventory/software'

function swTypeLabel(v: any): string {
  if (v === 0 || v === '0') return 'Unidentified'
  if (v === 1 || v === '1') return 'Commercial'
  if (v === 2 || v === '2') return 'Non-commercial'
  return String(v ?? '')
}

function usageLabel(v: any): string {
  if (v === true || v === 1) return 'Yes'
  if (v === false || v === 0) return 'No'
  return String(v ?? '')
}

const columns: ColumnDef[] = [
  { key: 'software_name', label: 'Name', format: 'bold' },
  { key: 'software_version', label: 'Version' },
  { key: 'display_name', label: 'Display Name' },
  { key: 'manufacturer_name', label: 'Manufacturer' },
  { key: 'sw_category_name', label: 'Category', badge: () => badge.zinc },
  { key: 'sw_type', label: 'Type', badge: () => badge.zinc, transform: swTypeLabel },
  { key: 'managed_installations', label: 'Managed', format: 'number' },
  { key: 'network_installations', label: 'Network', format: 'number' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'manufacturer_name', label: 'Manufacturer' },
  { key: 'sw_category_name', label: 'Category' },
  { key: 'sw_type', label: 'Type' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'software_name', label: 'Software Name' },
  { key: 'software_version', label: 'Software Version' },
  { key: 'display_name', label: 'Display Name' },
  { key: 'manufacturer_name', label: 'Manufacturer' },
  { key: 'manufacturer_id', label: 'Manufacturer ID', mono: true },
  { key: 'sw_category_name', label: 'Category' },
  { key: 'sw_type', label: 'Type', transform: swTypeLabel },
  { key: 'sw_family', label: 'Family' },
  { key: 'installed_format', label: 'Installed Format' },
  { key: 'is_usage_prohibited', label: 'Usage Prohibited', transform: usageLabel },
  { key: 'managed_installations', label: 'Managed Installations', transform: (v) => v != null ? String(v) : '' },
  { key: 'network_installations', label: 'Network Installations', transform: (v) => v != null ? String(v) : '' },
  { key: 'managed_sw_id', label: 'Managed SW ID', mono: true },
  { key: 'compliant_status', label: 'Compliant Status' },
  { key: 'total_copies', label: 'Total Copies', transform: (v) => v != null ? String(v) : '' },
  { key: 'remaining_copies', label: 'Remaining Copies', transform: (v) => v != null ? String(v) : '' },
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
    export-filename="meec-software.csv"
    drawer-title-key="software_name"
    title="Software"
  />
</template>
