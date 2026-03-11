<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/meec/inventory/installed-software'

function swTypeLabel(v: any): string {
  if (v === 0 || v === '0') return 'Unidentified'
  if (v === 1 || v === '1') return 'Commercial'
  if (v === 2 || v === '2') return 'Non-commercial'
  return String(v)
}

const columns: ColumnDef[] = [
  { key: 'software_name', label: 'Name', format: 'bold' },
  { key: 'software_version', label: 'Version' },
  { key: 'display_name', label: 'Display Name' },
  { key: 'manufacturer_name', label: 'Manufacturer' },
  { key: 'computer_resource_id', label: 'Computer', format: 'mono' },
  { key: 'architecture', label: 'Arch', badge: () => badge.zinc },
  { key: 'sw_category_name', label: 'Category', badge: () => badge.zinc },
  { key: 'sw_type', label: 'Type', badge: () => badge.zinc, transform: swTypeLabel },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'manufacturer_name', label: 'Manufacturer' },
  { key: 'architecture', label: 'Architecture' },
  { key: 'sw_category_name', label: 'Category' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'computer_resource_id', label: 'Computer', mono: true },
  { key: 'software_id', label: 'Software ID' },
  { key: 'software_name', label: 'Name' },
  { key: 'software_version', label: 'Version' },
  { key: 'display_name', label: 'Display Name' },
  { key: 'manufacturer_name', label: 'Manufacturer' },
  { key: 'architecture', label: 'Architecture' },
  { key: 'location', label: 'Location' },
  { key: 'sw_category_name', label: 'Category' },
  { key: 'sw_type', label: 'Type', transform: swTypeLabel },
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
    export-filename="meec-installed-software.csv"
    drawer-title-key="software_name"
    title="Installed Software"
  />
</template>
