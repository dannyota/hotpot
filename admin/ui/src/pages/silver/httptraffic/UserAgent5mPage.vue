<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/silver/httptraffic/user-agent-5m'

const methodBadge = badgeColors({
  GET: badge.blue,
  POST: badge.emerald,
  PUT: badge.amber,
  DELETE: badge.red,
})

const columns: ColumnDef[] = [
  { key: 'user_agent', label: 'User Agent' },
  { key: 'uri', label: 'URI', format: 'mono' },
  { key: 'method', badge: methodBadge },
  { key: 'ua_family', label: 'Family', badge: () => badge.zinc },
  { key: 'request_count', label: 'Requests', format: 'number' },
  { key: 'is_mapped', label: 'Mapped', format: 'bool' },
  { key: 'window_start', label: 'Window Start', format: 'date' },
  { key: 'window_end', label: 'Window End', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'ua_family', label: 'Family' },
  { key: 'is_mapped', label: 'Mapped', bool: true },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'endpoint_id', label: 'Endpoint ID', mono: true },
  { key: 'source_id', label: 'Source ID', mono: true },
  { key: 'user_agent', label: 'User Agent' },
  { key: 'uri', label: 'URI', mono: true },
  { key: 'method', label: 'Method' },
  { key: 'ua_family', label: 'Family' },
  { key: 'request_count', label: 'Requests' },
  { key: 'is_mapped', label: 'Mapped', format: 'bool' },
  { key: 'window_start', label: 'Window Start', format: 'date' },
  { key: 'window_end', label: 'Window End', format: 'date' },
  { key: 'collected_at', label: 'Last Seen', format: 'date' },
  { key: 'first_collected_at', label: 'First Seen', format: 'date' },
  { key: 'normalized_at', label: 'Normalized', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-window_start"
    search-key="name"
    search-placeholder="Search by user agent..."
    export-filename="silver-user-agent-5m.csv"
    drawer-title-key="user_agent"
    title="User Agent 5m"
  />
</template>
