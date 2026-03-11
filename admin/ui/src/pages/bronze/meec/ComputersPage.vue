<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/meec/inventory/computers'

function liveStatusLabel(v: any): string {
  switch (v) {
    case 1: return 'Live'
    case 2: return 'Down'
    case 3: return 'Unknown'
    default: return String(v ?? '')
  }
}

function liveStatusBadge(v: any): string {
  switch (v) {
    case 1: return badge.emerald
    case 2: return badge.red
    case 3: return badge.amber
    default: return badge.zinc
  }
}

function installStatusLabel(v: any): string {
  switch (v) {
    case 21: return 'Yet to Install'
    case 22: return 'Installed'
    case 23: return 'Uninstalled'
    default: return String(v ?? '')
  }
}

function installStatusBadge(v: any): string {
  switch (v) {
    case 22: return badge.emerald
    case 23: return badge.red
    case 21: return badge.amber
    default: return badge.zinc
  }
}

const columns: ColumnDef[] = [
  { key: 'resource_name', label: 'Name', format: 'bold' },
  { key: 'fqdn_name', label: 'FQDN' },
  { key: 'domain_netbios_name', label: 'Domain' },
  { key: 'ip_address', label: 'IP', format: 'mono' },
  { key: 'os_name', label: 'OS' },
  { key: 'os_platform_name', label: 'Platform', badge: () => badge.zinc },
  { key: 'agent_version', label: 'Agent' },
  { key: 'computer_live_status', label: 'Status', badge: liveStatusBadge, transform: liveStatusLabel },
  { key: 'installation_status', label: 'Install', badge: installStatusBadge, transform: installStatusLabel },
  { key: 'branch_office_name', label: 'Branch' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'os_platform_name', label: 'Platform' },
  { key: 'computer_live_status', label: 'Status' },
  { key: 'installation_status', label: 'Install' },
  { key: 'branch_office_name', label: 'Branch' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'resource_name', label: 'Name' },
  { key: 'fqdn_name', label: 'FQDN' },
  { key: 'domain_netbios_name', label: 'Domain' },
  { key: 'ip_address', label: 'IP Address', mono: true },
  { key: 'mac_address', label: 'MAC Address', mono: true },
  { key: 'os_name', label: 'OS' },
  { key: 'os_platform_name', label: 'Platform' },
  { key: 'os_version', label: 'OS Version' },
  { key: 'service_pack', label: 'Service Pack' },
  { key: 'agent_version', label: 'Agent Version' },
  { key: 'computer_live_status', label: 'Live Status', transform: liveStatusLabel },
  { key: 'installation_status', label: 'Install Status', transform: installStatusLabel },
  { key: 'managed_status', label: 'Managed Status' },
  { key: 'branch_office_name', label: 'Branch Office' },
  { key: 'owner', label: 'Owner' },
  { key: 'owner_email_id', label: 'Owner Email' },
  { key: 'description', label: 'Description' },
  { key: 'location', label: 'Location' },
  { key: 'customer_name', label: 'Customer' },
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
    search-key="resource_name"
    search-placeholder="Search by name..."
    export-filename="meec-computers.csv"
    drawer-title-key="resource_name"
    title="Computers"
  />
</template>
