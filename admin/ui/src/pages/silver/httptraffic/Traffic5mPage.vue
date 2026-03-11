<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/silver/httptraffic/traffic-5m'

const methodBadge = badgeColors({
  GET: badge.blue,
  POST: badge.emerald,
  PUT: badge.amber,
  DELETE: badge.red,
  PATCH: badge.purple,
})

const accessLevelBadge = badgeColors({
  PUBLIC: badge.emerald,
  PROTECTED: badge.amber,
  PRIVATE: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'uri', label: 'URI', format: 'mono' },
  { key: 'method', badge: methodBadge },
  { key: 'status_code', label: 'Status', format: 'number' },
  { key: 'request_count', label: 'Requests', format: 'number' },
  { key: 'avg_request_time', label: 'Avg Time', transform: (v) => v != null ? `${Number(v).toFixed(1)}ms` : '' },
  { key: 'unique_client_count', label: 'Clients', format: 'number' },
  { key: 'is_mapped', label: 'Mapped', format: 'bool' },
  { key: 'access_level', label: 'Access', badge: accessLevelBadge },
  { key: 'service', badge: () => badge.zinc },
  { key: 'window_start', label: 'Window Start', format: 'date' },
  { key: 'window_end', label: 'Window End', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'method' },
  { key: 'is_mapped', label: 'Mapped', bool: true },
  { key: 'service' },
  { key: 'access_level', label: 'Access Level' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'endpoint_id', label: 'Endpoint ID', mono: true },
  { key: 'source_id', label: 'Source ID', mono: true },
  { key: 'uri', label: 'URI', mono: true },
  { key: 'method', label: 'Method' },
  { key: 'status_code', label: 'Status Code' },
  { key: 'request_count', label: 'Request Count' },
  { key: 'avg_request_time', label: 'Avg Request Time', transform: (v) => v != null ? `${v} ms` : '' },
  { key: 'max_request_time', label: 'Max Request Time', transform: (v) => v != null ? `${v} ms` : '' },
  { key: 'total_body_bytes_sent', label: 'Total Body Bytes Sent' },
  { key: 'unique_client_count', label: 'Unique Clients' },
  { key: 'is_mapped', label: 'Mapped', format: 'bool' },
  { key: 'is_method_mismatch', label: 'Method Mismatch', format: 'bool' },
  { key: 'is_scanner_detected', label: 'Scanner Detected', format: 'bool' },
  { key: 'is_lfi_detected', label: 'LFI Detected', format: 'bool' },
  { key: 'is_sqli_detected', label: 'SQLi Detected', format: 'bool' },
  { key: 'is_rce_detected', label: 'RCE Detected', format: 'bool' },
  { key: 'is_xss_detected', label: 'XSS Detected', format: 'bool' },
  { key: 'is_ssrf_detected', label: 'SSRF Detected', format: 'bool' },
  { key: 'access_level', label: 'Access Level' },
  { key: 'service', label: 'Service' },
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
    search-key="q"
    search-placeholder="Search by URI..."
    export-filename="silver-traffic-5m.csv"
    drawer-title-key="uri"
    title="Traffic 5m"
  />
</template>
