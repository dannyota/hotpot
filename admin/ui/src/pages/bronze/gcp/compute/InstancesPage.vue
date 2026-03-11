<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import HGroupedStats from '@/components/app/HGroupedStats.vue'
import { shortPath, badge, badgeColors, dot } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/gcp/compute/instances'

const statusBadge = badgeColors({
  RUNNING: badge.emerald,
  TERMINATED: badge.zinc,
  STOPPED: badge.zinc,
  SUSPENDED: badge.amber,
  STAGING: badge.blue,
})

const statsDots: Record<string, string> = {
  RUNNING: dot.emeraldPulse,
  TERMINATED: dot.zinc,
  STOPPED: dot.zinc,
  SUSPENDED: dot.amberPulse,
  STAGING: dot.bluePulse,
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'zone', transform: shortPath },
  { key: 'machine_type', label: 'Machine Type', format: 'mono', transform: shortPath },
  { key: 'cpu_platform', label: 'CPU Platform' },
  { key: 'project_id', label: 'Project', format: 'mono' },
  { key: 'deletion_protection', label: 'Delete Protection', format: 'bool' },
  { key: 'creation_timestamp', label: 'Created', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'instance_type', label: 'Type' },
  { key: 'status' },
  { key: 'zone' },
  { key: 'machine_type', label: 'Machine Type' },
  { key: 'cpu_platform', label: 'CPU' },
  { key: 'project_id', label: 'Project' },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
  { key: 'zone', label: 'Zone' },
  { key: 'machine_type', label: 'Machine Type', transform: shortPath },
  { key: 'cpu_platform', label: 'CPU Platform' },
  { key: 'hostname', label: 'Hostname' },
  { key: 'description', label: 'Description' },
  { key: 'deletion_protection', label: 'Delete Protection', format: 'bool' },
  { key: 'can_ip_forward', label: 'Can IP Forward', format: 'bool' },
  { key: 'scheduling_json', label: 'Scheduling' },
  { key: 'creation_timestamp', label: 'Created' },
  { key: 'last_start_timestamp', label: 'Last Start' },
  { key: 'last_stop_timestamp', label: 'Last Stop' },
  { key: 'project_id', label: 'Project ID', mono: true },
  { key: 'self_link', label: 'Self Link', mono: true },
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
    :detail-route="(row) => `/bronze/gcp/compute/instances/${row.id}`"
    default-sort="-creation_timestamp"
    search-key="q"
    search-placeholder="Search by name..."
    export-filename="gcp-compute-instances.csv"
    drawer-title-key="name"
    title="Compute Instances"
    stats
  >
    <template #title="{ stats, meta, loading }">
      <HGroupedStats
        v-if="Object.keys(stats).length > 0"
        :stats="stats"
        :total="meta.total"
        :dot-colors="statsDots"
      />
      <h1 v-else class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
        {{ loading ? 'Compute Instances' : 'No Instances' }}
      </h1>
    </template>
  </HTablePage>
</template>
