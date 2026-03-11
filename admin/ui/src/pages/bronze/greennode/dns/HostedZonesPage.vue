<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/dns/hosted-zones'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
})

const typeBadge = badgeColors({
  PUBLIC: badge.blue,
  PRIVATE: badge.amber,
})

const columns: ColumnDef[] = [
  { key: 'domain_name', label: 'Domain Name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'type', badge: typeBadge },
  { key: 'description' },
  { key: 'count_records', label: 'Records', format: 'number' },
  { key: 'portal_user_id', label: 'Portal User', format: 'mono' },
  { key: 'created_at_api', label: 'Created (API)', format: 'date' },
  { key: 'updated_at_api', label: 'Updated (API)', format: 'date' },
  { key: 'deleted_at_api', label: 'Deleted (API)', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'type' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'domain_name', label: 'Domain Name' },
  { key: 'status', label: 'Status' },
  { key: 'type', label: 'Type' },
  { key: 'description', label: 'Description' },
  { key: 'count_records', label: 'Records' },
  { key: 'portal_user_id', label: 'Portal User ID', mono: true },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'created_at_api', label: 'Created (API)', format: 'date' },
  { key: 'updated_at_api', label: 'Updated (API)', format: 'date' },
  { key: 'deleted_at_api', label: 'Deleted (API)', format: 'date' },
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
    search-key="domain_name"
    search-placeholder="Search by domain name..."
    export-filename="greennode-dns-hosted-zones.csv"
    drawer-title-key="domain_name"
    title="Hosted Zones"
  />
</template>
