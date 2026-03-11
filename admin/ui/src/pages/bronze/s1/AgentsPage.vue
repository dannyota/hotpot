<script setup lang="ts">
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors, dot } from '@/composables/formatting'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/s1/agents'

const osTypeBadge = badgeColors({
  WINDOWS: badge.blue,
  LINUX: badge.amber,
  MACOS: badge.purple,
})

const networkBadge = badgeColors({
  CONNECTED: badge.emerald,
  DISCONNECTED: badge.zinc,
})

const columns: ColumnDef[] = [
  { key: 'computer_name', label: 'Name', format: 'bold' },
  { key: 'os_type', label: 'OS Type', badge: osTypeBadge },
  { key: 'os_name', label: 'OS' },
  { key: 'agent_version', label: 'Version' },
  { key: 'is_active', label: 'Active', badge: (v) => v ? badge.emerald : badge.amber, transform: (v) => v ? 'Yes' : 'No' },
  { key: 'is_infected', label: 'Infected', badge: (v) => v ? badge.red : badge.emerald, transform: (v) => v ? 'Yes' : 'No' },
  { key: 'network_status', label: 'Network', badge: networkBadge },
  { key: 'site_name', label: 'Site' },
  { key: 'last_active_date', label: 'Last Active', format: 'date' },
  { key: 'first_collected_at', format: 'date' },
  { key: 'collected_at', format: 'relative' },
]

const filters: FilterDef[] = [
  { key: 'os_type', label: 'OS Type' },
  { key: 'site_name', label: 'Site' },
  { key: 'network_status', label: 'Network' },
  { key: 'is_active', label: 'Active', bool: true },
  { key: 'is_infected', label: 'Infected', bool: true },
]

const drawerFields: DrawerFieldDef[] = [
  { key: 'resource_id', label: 'Resource ID', mono: true },
  { key: 'computer_name', label: 'Computer Name' },
  { key: 'external_ip', label: 'External IP' },
  { key: 'os_type', label: 'OS Type' },
  { key: 'os_name', label: 'OS Name' },
  { key: 'os_revision', label: 'OS Revision' },
  { key: 'os_arch', label: 'OS Arch' },
  { key: 'agent_version', label: 'Agent Version' },
  { key: 'is_active', label: 'Active', format: 'bool' },
  { key: 'is_infected', label: 'Infected', format: 'bool' },
  { key: 'is_decommissioned', label: 'Decommissioned', format: 'bool' },
  { key: 'machine_type', label: 'Machine Type' },
  { key: 'domain', label: 'Domain' },
  { key: 'network_status', label: 'Network Status' },
  { key: 'site_name', label: 'Site' },
  { key: 'account_name', label: 'Account' },
  { key: 'group_name', label: 'Group' },
  { key: 'last_active_date', label: 'Last Active', format: 'date' },
  { key: 'registered_at', label: 'Registered', format: 'date' },
  { key: 'active_threats', label: 'Active Threats', transform: (v) => v != null ? String(v) : '' },
  { key: 'cpu_count', label: 'CPU Count', transform: (v) => v != null ? String(v) : '' },
  { key: 'core_count', label: 'Core Count', transform: (v) => v != null ? String(v) : '' },
  { key: 'total_memory', label: 'Total Memory', transform: (v) => v != null ? String(v) : '' },
  { key: 'model_name', label: 'Model' },
  { key: 'serial_number', label: 'Serial Number' },
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
    default-sort="-last_active_date"
    search-key="computer_name"
    search-placeholder="Search by name..."
    export-filename="s1-agents.csv"
    drawer-title-key="computer_name"
    title="Agents"
    stats
  >
    <template #title="{ stats, loading }">
      <div v-if="stats.total > 0" class="flex items-center gap-2 text-sm flex-wrap">
        <span class="inline-flex items-center gap-1">
          <span class="w-1.5 h-1.5 rounded-full" :class="dot.emeraldPulse" />
          <span class="font-semibold text-zinc-900 dark:text-zinc-100">{{ stats.active?.toLocaleString() }}</span>
          <span class="text-zinc-500 dark:text-zinc-400">active</span>
        </span>
        <span class="text-zinc-300 dark:text-zinc-600">&bull;</span>
        <span class="inline-flex items-center gap-1">
          <span class="w-1.5 h-1.5 rounded-full" :class="dot.zinc" />
          <span class="font-semibold text-zinc-900 dark:text-zinc-100">{{ stats.inactive?.toLocaleString() }}</span>
          <span class="text-zinc-500 dark:text-zinc-400">inactive</span>
        </span>
        <template v-if="stats.infected">
          <span class="text-zinc-300 dark:text-zinc-600">&bull;</span>
          <span class="inline-flex items-center gap-1">
            <span class="w-1.5 h-1.5 rounded-full" :class="dot.redPulse" />
            <span class="font-semibold text-zinc-900 dark:text-zinc-100">{{ stats.infected?.toLocaleString() }}</span>
            <span class="text-zinc-500 dark:text-zinc-400">infected</span>
          </span>
        </template>
        <span class="text-zinc-300 dark:text-zinc-600">&bull;</span>
        <span class="inline-flex items-center gap-1">
          <span class="w-1.5 h-1.5 rounded-full" :class="dot.emeraldPulse" />
          <span class="font-semibold text-zinc-900 dark:text-zinc-100">{{ stats.connected?.toLocaleString() }}</span>
          <span class="text-zinc-500 dark:text-zinc-400">connected</span>
        </span>
        <span class="text-zinc-300 dark:text-zinc-600">&bull;</span>
        <span class="text-zinc-500 dark:text-zinc-400">
          TOTAL: <span class="font-semibold text-zinc-900 dark:text-zinc-100">{{ stats.total?.toLocaleString() }}</span>
        </span>
      </div>
      <h1 v-else class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Agents</h1>
    </template>
  </HTablePage>
</template>
