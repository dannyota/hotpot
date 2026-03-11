<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/gold/httpmonitor/anomalies'

const severityBadge = badgeColors({
  CRITICAL: badge.red,
  HIGH: badge.amber,
  MEDIUM: badge.amber,
  LOW: badge.blue,
  INFO: badge.zinc,
})

const methodBadge = badgeColors({
  GET: badge.blue,
  POST: badge.emerald,
  PUT: badge.amber,
  DELETE: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'anomaly_type', label: 'Type', badge: () => badge.zinc },
  { key: 'severity', badge: severityBadge },
  { key: 'uri', label: 'URI', format: 'mono' },
  { key: 'method', badge: methodBadge },
  { key: 'baseline_value', label: 'Baseline', transform: (v) => v != null ? Number(v).toFixed(1) : '' },
  { key: 'actual_value', label: 'Actual', transform: (v) => v != null ? Number(v).toFixed(1) : '' },
  { key: 'deviation', transform: (v) => v != null ? `${Number(v).toFixed(2)}x` : '' },
  { key: 'description' },
  { key: 'window_start', label: 'Window Start', format: 'date' },
  { key: 'window_end', label: 'Window End', format: 'date' },
  { key: 'first_detected_at', label: 'First Detected', format: 'date' },
  { key: 'detected_at', label: 'Detected', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'anomaly_type', label: 'Type' },
  { key: 'severity' },
  { key: 'method' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'endpoint_id', label: 'Endpoint ID', mono: true },
  { key: 'source_id', label: 'Source ID', mono: true },
  { key: 'anomaly_type', label: 'Anomaly Type' },
  { key: 'severity', label: 'Severity' },
  { key: 'uri', label: 'URI', mono: true },
  { key: 'method', label: 'Method' },
  { key: 'baseline_value', label: 'Baseline Value' },
  { key: 'actual_value', label: 'Actual Value' },
  { key: 'deviation', label: 'Deviation' },
  { key: 'description', label: 'Description' },
  { key: 'window_start', label: 'Window Start', format: 'date' },
  { key: 'window_end', label: 'Window End', format: 'date' },
  { key: 'detected_at', label: 'Detected', format: 'date' },
  { key: 'first_detected_at', label: 'First Detected', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-detected_at"
    search-key="q"
    search-placeholder="Search by URI..."
    export-filename="gold-httpmonitor-anomalies.csv"
    drawer-title-key="anomaly_type"
    title="Anomalies"
  />
</template>
