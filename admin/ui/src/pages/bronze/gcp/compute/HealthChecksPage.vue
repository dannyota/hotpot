<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/health-checks'

const typeBadge = badgeColors({
  TCP: badge.blue,
  HTTP: badge.emerald,
  HTTPS: badge.emerald,
  SSL: badge.amber,
  GRPC: badge.purple,
})

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'type', badge: typeBadge },
  { key: 'check_interval_sec', label: 'Interval', format: 'number', transform: (v) => v != null ? `${v}s` : '' },
  { key: 'timeout_sec', label: 'Timeout', format: 'number', transform: (v) => v != null ? `${v}s` : '' },
  { key: 'healthy_threshold', label: 'Healthy', format: 'number' },
  { key: 'unhealthy_threshold', label: 'Unhealthy', format: 'number' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'type' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'type', label: 'Type' },
  { key: 'check_interval_sec', label: 'Check Interval', transform: (v) => v != null ? `${v}s` : '' },
  { key: 'timeout_sec', label: 'Timeout', transform: (v) => v != null ? `${v}s` : '' },
  { key: 'healthy_threshold', label: 'Healthy Threshold' },
  { key: 'unhealthy_threshold', label: 'Unhealthy Threshold' },
  { key: 'region', label: 'Region' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'self_link', label: 'Self Link', mono: true },
  { key: 'creation_timestamp', label: 'Created' },
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
    default-sort="-creation_timestamp"
    export-filename="gcp-compute-health-checks.csv"
    title="Health Checks"
  />
</template>
