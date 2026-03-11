<script setup lang="ts">
import { ref } from 'vue'
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/network/subnets'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
})

// --- VPC lookup (network_id -> VPC info) ---
const vpcMap = ref<Record<string, { name: string; cidr: string }>>({})
const vpcFetched = new Set<string>()

async function resolveVPCs(networkIds: string[]) {
  const missing = networkIds.filter(id => id && !vpcFetched.has(id))
  if (missing.length === 0) return
  for (const id of missing) vpcFetched.add(id)
  try {
    const res = await window.fetch(`/api/v1/bronze/greennode/network/vpcs?size=${missing.length}&filter[id]=${encodeURIComponent(missing.join(','))}`)
    if (!res.ok) return
    const json = await res.json()
    const updated = { ...vpcMap.value }
    for (const vpc of json.data) {
      updated[vpc.id] = { name: vpc.name, cidr: vpc.cidr }
    }
    vpcMap.value = updated
  } catch { /* ignore */ }
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'cidr', label: 'CIDR', format: 'mono' },
  { key: 'status', badge: statusBadge },
  { key: 'network_id', label: 'Network' },
  { key: 'zone_id', label: 'Zone' },
  { key: 'region' },
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
  { key: 'network_id', label: 'Network ID', mono: true },
  { key: 'network_id', label: 'Network Name', transform: (v) => vpcMap.value[v]?.name ?? '' },
  { key: 'network_id', label: 'Network CIDR', mono: true, transform: (v) => vpcMap.value[v]?.cidr ?? '' },
  { key: 'zone_id', label: 'Zone' },
  { key: 'region', label: 'Region' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'route_table_id', label: 'Route Table ID', mono: true },
  { key: 'interface_acl_policy_name', label: 'ACL Policy' },
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
    export-filename="greennode-subnets.csv"
    title="Subnets"
    @data="(rows: any[]) => { const ids = [...new Set(rows.map((r: any) => r.network_id).filter(Boolean))]; resolveVPCs(ids) }"
  >
    <template #cell-network_id="{ value }">
      <span v-if="vpcMap[value]" class="text-sm">
        {{ vpcMap[value].name }}
        <span class="font-mono text-xs text-zinc-400 dark:text-zinc-500 ml-1">{{ vpcMap[value].cidr }}</span>
      </span>
      <span v-else class="font-mono text-xs text-zinc-500 dark:text-zinc-400">{{ value }}</span>
    </template>
  </HTablePage>
</template>
