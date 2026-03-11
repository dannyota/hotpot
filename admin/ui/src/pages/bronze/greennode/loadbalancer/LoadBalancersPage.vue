<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import HJsonPeek from '@/components/app/HJsonPeek.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/loadbalancer/lbs'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
  ERROR: badge.red,
})

function nodeCount(nodes: any): number {
  if (Array.isArray(nodes)) return nodes.length
  return 0
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'address', format: 'mono' },
  { key: 'type' },
  { key: 'location' },
  { key: 'region' },
  { key: 'total_nodes', label: 'Nodes', format: 'number' },
  { key: 'nodes_json', label: 'Node Details', sortable: false },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'region' },
  { key: 'location' },
  { key: 'type' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'address', label: 'Address', mono: true },
  { key: 'type', label: 'Type' },
  { key: 'location', label: 'Location' },
  { key: 'region', label: 'Region' },
  { key: 'total_nodes', label: 'Total Nodes' },
  { key: 'nodes_json', label: 'Nodes' },
  { key: 'private_subnet_cidr', label: 'Private Subnet CIDR', mono: true },
  { key: 'package_id', label: 'Package ID', mono: true },
  { key: 'created_at_api', label: 'Created', format: 'date' },
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
    export-filename="greennode-loadbalancers.csv"
    title="Load Balancers"
  >
    <template #cell-nodes_json="{ value, row, showJson }">
      <HJsonPeek
        :label="`${nodeCount(value)} nodes`"
        :data="value"
        :title="`Nodes — ${row.name}`"
        @click="showJson"
      />
    </template>
  </HTablePage>
</template>
