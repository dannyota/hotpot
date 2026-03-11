<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/network/endpoints'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
  ERROR: badge.red,
})

const billingBadge = badgeColors({
  NORMAL: badge.emerald,
  FROZEN: badge.amber,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'billing_status', label: 'Billing', badge: billingBadge },
  { key: 'endpoint_type', label: 'Type' },
  { key: 'ipv4_address', label: 'IPv4', format: 'mono' },
  { key: 'service_name', label: 'Service' },
  { key: 'vpc_name', label: 'VPC' },
  { key: 'region' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'billing_status', label: 'Billing' },
  { key: 'endpoint_type', label: 'Type' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'billing_status', label: 'Billing Status' },
  { key: 'endpoint_type', label: 'Endpoint Type' },
  { key: 'ipv4_address', label: 'IPv4 Address', mono: true },
  { key: 'endpoint_url', label: 'Endpoint URL', mono: true },
  { key: 'endpoint_service_id', label: 'Endpoint Service ID', mono: true },
  { key: 'service_name', label: 'Service' },
  { key: 'service_endpoint_type', label: 'Service Endpoint Type' },
  { key: 'category_name', label: 'Category' },
  { key: 'package_name', label: 'Package' },
  { key: 'version', label: 'Version' },
  { key: 'description', label: 'Description' },
  { key: 'vpc_id', label: 'VPC ID', mono: true },
  { key: 'vpc_name', label: 'VPC Name' },
  { key: 'zone_uuid', label: 'Zone', mono: true },
  { key: 'region', label: 'Region' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'created_at', label: 'Created', format: 'date' },
  { key: 'updated_at', label: 'Updated', format: 'date' },
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
    export-filename="greennode-endpoints.csv"
    title="Network Endpoints"
  />
</template>
