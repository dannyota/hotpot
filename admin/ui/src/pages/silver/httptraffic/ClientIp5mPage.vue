<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/silver/httptraffic/client-ip-5m'

const methodBadge = badgeColors({
  GET: badge.blue,
  POST: badge.emerald,
  PUT: badge.amber,
  DELETE: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'client_ip', label: 'Client IP', format: 'mono' },
  { key: 'uri', label: 'URI', format: 'mono' },
  { key: 'method', badge: methodBadge },
  { key: 'country_code', label: 'Country', badge: () => badge.zinc },
  { key: 'org_name', label: 'Organization' },
  { key: 'is_internal', label: 'Internal', format: 'bool' },
  { key: 'request_count', label: 'Requests', format: 'number' },
  { key: 'is_mapped', label: 'Mapped', format: 'bool' },
  { key: 'window_start', label: 'Window Start', format: 'date' },
  { key: 'window_end', label: 'Window End', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'country_code', label: 'Country' },
  { key: 'org_name', label: 'Organization' },
  { key: 'is_internal', label: 'Internal', bool: true },
  { key: 'is_mapped', label: 'Mapped', bool: true },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'endpoint_id', label: 'Endpoint ID', mono: true },
  { key: 'source_id', label: 'Source ID', mono: true },
  { key: 'client_ip', label: 'Client IP', mono: true },
  { key: 'uri', label: 'URI', mono: true },
  { key: 'method', label: 'Method' },
  { key: 'country_code', label: 'Country Code' },
  { key: 'country_name', label: 'Country Name' },
  { key: 'asn', label: 'ASN' },
  { key: 'org_name', label: 'Organization' },
  { key: 'as_domain', label: 'AS Domain' },
  { key: 'asn_type', label: 'ASN Type' },
  { key: 'is_internal', label: 'Internal', format: 'bool' },
  { key: 'request_count', label: 'Request Count' },
  { key: 'is_mapped', label: 'Mapped', format: 'bool' },
  { key: 'window_start', label: 'Window Start', format: 'date' },
  { key: 'window_end', label: 'Window End', format: 'date' },
  { key: 'collected_at', label: 'Collected At', format: 'date' },
  { key: 'first_collected_at', label: 'First Collected At', format: 'date' },
  { key: 'normalized_at', label: 'Normalized At', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-window_start"
    search-key="client_ip"
    search-placeholder="Search by IP..."
    export-filename="silver-client-ip-5m.csv"
    drawer-title-key="client_ip"
    title="Client IP 5m"
  />
</template>
