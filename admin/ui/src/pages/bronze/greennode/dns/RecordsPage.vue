<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/dns/records'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
  ERROR: badge.red,
})

const typeBadge = badgeColors({
  A: badge.blue,
  CNAME: badge.amber,
}, badge.zinc)

const routingPolicyBadge = badgeColors({
  SIMPLE: badge.zinc,
}, badge.zinc)

const columns: ColumnDef[] = [
  { key: 'sub_domain', label: 'Sub Domain', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'type', badge: typeBadge },
  { key: 'routing_policy', label: 'Routing Policy', badge: routingPolicyBadge },
  { key: 'ttl', label: 'TTL' },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'updated_at_api', label: 'Updated', format: 'date' },
  { key: 'deleted_at_api', label: 'Deleted', format: 'date' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'type' },
  { key: 'routing_policy', label: 'Routing Policy' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'id', label: 'ID', mono: true },
  { key: 'record_id', label: 'Record ID', mono: true },
  { key: 'sub_domain', label: 'Sub Domain' },
  { key: 'status', label: 'Status' },
  { key: 'type', label: 'Type' },
  { key: 'routing_policy', label: 'Routing Policy' },
  { key: 'ttl', label: 'TTL' },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'updated_at_api', label: 'Updated', format: 'date' },
  { key: 'deleted_at_api', label: 'Deleted', format: 'date' },
]
</script>

<template>
  <HTablePage
    :endpoint="ENDPOINT"
    :columns="columns"
    :drawer-fields="drawerFields"
    :filters="filters"
    default-sort="-id"
    search-key="q"
    search-placeholder="Search by sub domain..."
    export-filename="greennode-dns-records.csv"
    drawer-title-key="sub_domain"
    title="DNS Records"
  >
    <template #cell-ttl="{ value }">
      <span v-if="value != null" class="tabular-nums">{{ value }}s</span>
    </template>
  </HTablePage>
</template>
