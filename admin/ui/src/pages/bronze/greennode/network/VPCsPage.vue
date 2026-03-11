<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import HJsonPeek from '@/components/app/HJsonPeek.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/network/vpcs'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
})

function elasticIpCount(ips: any): number {
  if (Array.isArray(ips)) return ips.length
  return 0
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'cidr', label: 'CIDR', format: 'mono' },
  { key: 'status', badge: statusBadge },
  { key: 'dns_status', label: 'DNS Status' },
  { key: 'zone_name', label: 'Zone' },
  { key: 'region' },
  { key: 'elastic_ips', label: 'Elastic IPs', sortable: false },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'region' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'cidr', label: 'CIDR', mono: true },
  { key: 'status', label: 'Status' },
  { key: 'dns_status', label: 'DNS Status' },
  { key: 'zone_name', label: 'Zone' },
  { key: 'region', label: 'Region' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'route_table_name', label: 'Route Table' },
  { key: 'dhcp_option_name', label: 'DHCP Option' },
  { key: 'elastic_ips', label: 'Elastic IPs' },
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
    export-filename="greennode-vpcs.csv"
    title="VPCs"
  >
    <template #cell-elastic_ips="{ value, row, showJson }">
      <HJsonPeek
        :label="`${elasticIpCount(value)} IPs`"
        :data="value"
        :title="`Elastic IPs — ${row.name}`"
        @click="showJson"
      />
    </template>
  </HTablePage>
</template>
