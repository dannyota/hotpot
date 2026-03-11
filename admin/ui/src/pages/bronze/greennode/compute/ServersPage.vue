<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import HGroupedStats from '@/components/app/HGroupedStats.vue'
import HJsonPeek from '@/components/app/HJsonPeek.vue'
import { badge, badgeColors, dot } from '@/composables/formatting'
import { useTimezone } from '@/composables/useTimezone'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const { formatDateTime } = useTimezone()
const ENDPOINT = '/api/v1/bronze/greennode/compute/servers'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
  SHUTOFF: badge.zinc,
  ERROR: badge.red,
  BUILD: badge.blue,
})

const statsDots: Record<string, string> = {
  ACTIVE: dot.emeraldPulse,
  SHUTOFF: dot.zinc,
  STOPPED: dot.zinc,
  ERROR: dot.red,
  BUILD: dot.bluePulse,
}

function formatMemory(mb: number): string {
  if (!mb) return ''
  if (mb >= 1024) return `${(mb / 1024).toFixed(mb % 1024 === 0 ? 0 : 1)} GB`
  return `${mb} MB`
}

function networkSummary(ifaces: any): { ip: string; extra: number } {
  const ips: string[] = []
  for (const group of [ifaces?.internal, ifaces?.external]) {
    if (Array.isArray(group)) {
      for (const iface of group) {
        if (iface.fixedIp) ips.push(iface.fixedIp)
      }
    }
  }
  return { ip: ips[0] ?? '', extra: Math.max(0, ips.length - 1) }
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'region' },
  { key: 'location' },
  { key: 'product' },
  { key: 'flavor_name', label: 'Flavor' },
  { key: 'flavor_cpu', label: 'CPU' },
  { key: 'flavor_memory', label: 'Memory' },
  { key: 'server_group_name', label: 'Server Group' },
  { key: 'image_type', label: 'Image' },
  { key: 'interfaces_json', label: 'Networks', sortable: false },
  { key: 'created_at_api', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'status' },
  { key: 'region' },
  { key: 'location' },
  { key: 'product' },
  { key: 'server_group_name', label: 'Server Group' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'region', label: 'Region' },
  { key: 'location', label: 'Location' },
  { key: 'product', label: 'Product' },
  { key: 'flavor_name', label: 'Flavor' },
  { key: 'flavor_cpu', label: 'CPU', transform: (v) => v ? `${v} vCPU` : '' },
  { key: 'flavor_memory', label: 'Memory', transform: (v) => formatMemory(v) },
  { key: 'server_group_name', label: 'Server Group' },
  { key: 'server_group_id', label: 'Server Group ID', mono: true },
  { key: 'image_type', label: 'Image Type' },
  { key: 'image_id', label: 'Image ID', mono: true },
  { key: 'image_version', label: 'Image Version' },
  { key: 'interfaces_json', label: 'Networks' },
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
    default-sort="-created_at_api"
    search-key="q"
    search-placeholder="Search by name or IP..."
    export-filename="greennode-servers.csv"
    drawer-title-key="name"
    title="Servers"
    stats
  >
    <template #title="{ stats, meta, loading }">
      <HGroupedStats
        v-if="Object.keys(stats).length > 0"
        :stats="stats"
        :total="meta.total"
        :dot-colors="statsDots"
        uppercase-breakdown
      />
      <h1 v-else class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
        {{ loading ? 'Servers' : 'No Servers' }}
      </h1>
    </template>

    <template #cell-flavor_cpu="{ value }">
      <span v-if="value" class="tabular-nums">{{ value }} vCPU</span>
    </template>

    <template #cell-flavor_memory="{ value }">
      <span v-if="value" class="tabular-nums">{{ formatMemory(value) }}</span>
    </template>

    <template #cell-image_type="{ row, showJson }">
      <HJsonPeek
        :label="row.image_type"
        :data="row.image_type ? { image_id: row.image_id, image_type: row.image_type, image_version: row.image_version } : null"
        :title="`Image — ${row.name}`"
        @click="showJson"
      />
    </template>

    <template #cell-interfaces_json="{ value, row, showJson }">
      <HJsonPeek
        :label="networkSummary(value).ip"
        :extra="networkSummary(value).extra"
        :data="value"
        :title="`Network — ${row.name}`"
        @click="showJson"
      />
    </template>
  </HTablePage>
</template>
