<script setup lang="ts">
import { ref } from 'vue'
import HTablePage from '@/components/app/HTablePage.vue'
import { badge, badgeColors } from '@/composables/formatting'
import { Copy, Check } from 'lucide-vue-next'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'

const ENDPOINT = '/api/v1/bronze/greennode/compute/ssh-keys'

const statusBadge = badgeColors({
  ACTIVE: badge.emerald,
})

function truncateKey(key: string): string {
  if (!key) return ''
  const parts = key.split(' ')
  if (parts.length >= 2) {
    const body = parts[1]
    return `${parts[0]} ${body.slice(0, 16)}...${body.slice(-8)}${parts[2] ? ' ' + parts[2] : ''}`
  }
  if (key.length > 40) return key.slice(0, 20) + '...' + key.slice(-12)
  return key
}

const copiedId = ref<string | null>(null)

async function copyKey(key: string, id: string) {
  try {
    await navigator.clipboard.writeText(key)
    copiedId.value = id
    setTimeout(() => { copiedId.value = null }, 1500)
  } catch { /* ignore */ }
}

const columns: ColumnDef[] = [
  { key: 'name', format: 'bold' },
  { key: 'status', badge: statusBadge },
  { key: 'pub_key', label: 'Public Key', sortable: false },
  { key: 'region' },
  { key: 'created_at_api', label: 'Created', format: 'date' },
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
  { key: 'status', label: 'Status' },
  { key: 'pub_key', label: 'Public Key', mono: true },
  { key: 'region', label: 'Region' },
  { key: 'project_id', label: 'Project ID', mono: true },
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
    export-filename="greennode-ssh-keys.csv"
    title="SSH Keys"
  >
    <template #cell-pub_key="{ value, row }">
      <div class="flex items-center gap-1.5 max-w-[300px]">
        <span class="font-mono text-xs text-zinc-500 dark:text-zinc-400 truncate">{{ truncateKey(value) }}</span>
        <button
          v-if="value"
          class="shrink-0 p-0.5 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300 transition-colors"
          title="Copy public key"
          @click.stop="copyKey(value, row.resource_id)"
        >
          <Check v-if="copiedId === row.resource_id" class="w-3.5 h-3.5 text-emerald-500" />
          <Copy v-else class="w-3.5 h-3.5" />
        </button>
      </div>
    </template>
  </HTablePage>
</template>
