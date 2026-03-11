<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useTableQuery } from '@/composables/useTableQuery'
import { exportCsv } from '@/composables/useExport'
import { useColumnChooser } from '@/composables/useColumnChooser'
import type { Column } from '@/components/app/HDataTable.vue'
import HDataTable from '@/components/app/HDataTable.vue'
import HMultiSelect from '@/components/app/HMultiSelect.vue'
import HColumnChooser from '@/components/app/HColumnChooser.vue'
import HRefreshButton from '@/components/app/HRefreshButton.vue'
import HDetailDrawer from '@/components/app/HDetailDrawer.vue'
import HJsonViewer from '@/components/app/HJsonViewer.vue'
import HJsonPeek from '@/components/app/HJsonPeek.vue'
import HDateTime from '@/components/app/HDateTime.vue'
import HRelativeTime from '@/components/app/HRelativeTime.vue'
import { Search, X, Download } from 'lucide-vue-next'
import { columnLabel } from '@/composables/columns'
import { useTimezone } from '@/composables/useTimezone'

const route = useRoute()
const { formatDateTime } = useTimezone()

const api = computed(() => (route.meta?.api as string) ?? '')
const title = computed(() => (route.meta?.label as string) ?? 'Data')

const {
  data, meta, filterOptions, loading, error, sort, filters,
  reload, setSort, setFilter, clearFilters, setPage, setPageSize,
  searchText, multiFilters, onSearchInput, onMultiFilterChange, hasActiveFilters, onClearAll,
} = useTableQuery<Record<string, any>>({
  endpoint: () => api.value,
  defaultSort: '',
})

// --- Column type detection ---
function isDateValue(v: unknown): boolean {
  if (typeof v !== 'string') return false
  return /^\d{4}-\d{2}-\d{2}[T ]/.test(v)
}

function isJsonValue(v: unknown): boolean {
  return v !== null && typeof v === 'object'
}

function isBoolValue(v: unknown): boolean {
  return typeof v === 'boolean'
}

// Columns always end with first_collected_at, collected_at.
const trailingCols = new Set(['first_collected_at', 'collected_at'])
const hiddenCols = computed(() => new Set((route.meta?.hideCols as string[]) ?? []))

const columns = computed<Column<Record<string, any>>[]>(() => {
  if (data.value.length === 0) return []
  const row = data.value[0]
  const keys = Object.keys(row).filter(k => !k.startsWith('edges') && !hiddenCols.value.has(k))
  const main = keys.filter(k => !trailingCols.has(k))
  const tail = ['first_collected_at', 'collected_at'].filter(k => keys.includes(k))
  return [...main, ...tail].map(k => ({
    key: k,
    label: columnLabel(k),
    sortable: !isJsonValue(row[k]),
  }))
})

const { visibleColumns, hiddenKeys, setVisible, resetAll } = useColumnChooser(api, columns)

// Track which columns are dates, json, or booleans based on first row.
const columnTypes = computed<Record<string, 'date' | 'json' | 'bool' | 'text'>>(() => {
  if (data.value.length === 0) return {}
  const row = data.value[0]
  const types: Record<string, 'date' | 'json' | 'bool' | 'text'> = {}
  for (const k of Object.keys(row)) {
    const v = row[k]
    if (isJsonValue(v)) types[k] = 'json'
    else if (isBoolValue(v)) types[k] = 'bool'
    else if (isDateValue(v)) types[k] = 'date'
    else types[k] = 'text'
  }
  return types
})

const availableFilterKeys = computed(() => Object.keys(filterOptions.value))

// --- JSON viewer ---
const jsonViewerData = ref<any>(null)
const jsonViewerTitle = ref('')

function showJson(data: any, title: string) {
  jsonViewerData.value = data
  jsonViewerTitle.value = title
}

// --- Detail drawer ---
const drawerRow = ref<Record<string, any> | null>(null)

function drawerFields(row: Record<string, any>) {
  return Object.entries(row)
    .filter(([k]) => !k.startsWith('edges'))
    .map(([k, v]) => ({
      label: columnLabel(k),
      value: isDateValue(v) ? formatDateTime(v as string) : v,
      mono: k.endsWith('_id') || k === 'id',
    }))
}

// --- Export ---
const exporting = ref(false)

async function onExport() {
  exporting.value = true
  try {
    await exportCsv(api.value, filters.value, sort.value, meta.value.total, `${title.value}.csv`)
  } finally {
    exporting.value = false
  }
}

// Close drawer when navigating between generic pages.
// useTableQuery restores saved filters internally and triggers auto-reload.
watch(api, (newApi) => {
  if (newApi) {
    drawerRow.value = null
  }
})

onMounted(() => {
  if (api.value) reload()
})
</script>

<template>
  <div class="p-6 space-y-2 max-w-full">
    <!-- Title + Export -->
    <div class="flex items-center justify-between gap-4">
      <h1 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">{{ title }}</h1>
      <div class="flex items-center gap-2">
        <HRefreshButton :loading="loading" />
        <HColumnChooser
          v-if="columns.length > 0"
          :columns="columns"
          :hidden-keys="hiddenKeys"
          @toggle="setVisible"
          @reset="resetAll"
        />
        <button
          v-if="meta.total > 0"
          :disabled="exporting"
          class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-zinc-600 dark:text-zinc-400 border border-zinc-200 dark:border-zinc-700 rounded-md hover:bg-zinc-50 dark:hover:bg-zinc-800 disabled:opacity-50 transition-colors shrink-0"
          @click="onExport"
        >
          <Download class="w-3.5 h-3.5" />
          {{ exporting ? 'Exporting...' : 'Export CSV' }}
        </button>
      </div>
    </div>

    <!-- Search + Filters -->
    <div class="flex items-center gap-2 flex-wrap">
      <div class="relative flex-1 max-w-xs">
        <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
        <input
          v-model="searchText"
          type="text"
          placeholder="Search by name..."
          class="w-full pl-8 pr-3 py-1.5 text-sm border border-zinc-200 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-900 text-zinc-900 dark:text-zinc-100 placeholder:text-zinc-400 focus:outline-none focus:ring-2 focus:ring-zinc-900/10 dark:focus:ring-zinc-100/10 transition-shadow"
          @input="onSearchInput"
        />
      </div>

      <HMultiSelect
        v-for="field in availableFilterKeys"
        :key="field"
        :label="field.replace(/_/g, ' ').replace(/\b\w/g, (c: string) => c.toUpperCase())"
        :options="filterOptions[field] ?? []"
        :model-value="multiFilters[field] ?? []"
        @update:model-value="onMultiFilterChange(field, $event)"
      />

      <button
        v-if="hasActiveFilters"
        class="inline-flex items-center gap-1 px-2 py-1.5 text-xs text-zinc-500 hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-200 transition-colors"
        @click="onClearAll"
      >
        <X class="w-3 h-3" />Clear
      </button>
    </div>

    <!-- Table -->
    <HDataTable
      :columns="visibleColumns"
      :data="data"
      :meta="meta"
      :sort="sort"
      :loading="loading"
      :error="error"
      @sort="setSort"
      @page="setPage"
      @page-size="setPageSize"
      @row-click="(row) => drawerRow = row"
    >
      <!-- Dynamic slots for each column based on detected type -->
      <template v-for="col in visibleColumns" :key="col.key" #[col.key]="{ value, row }">
        <!-- collected_at: relative time with tooltip -->
        <template v-if="col.key === 'collected_at'">
          <HRelativeTime :value="value" />
        </template>

        <!-- JSON: hover preview + click for full modal -->
        <template v-else-if="columnTypes[col.key] === 'json'">
          <HJsonPeek
            :data="value"
            :title="`${col.label} — ${row.name ?? row.id ?? ''}`"
            @click="showJson"
          />
        </template>

        <!-- Boolean: badge -->
        <template v-else-if="columnTypes[col.key] === 'bool'">
          <span
            class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full"
            :class="value ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400'"
          >{{ value ? 'Yes' : 'No' }}</span>
        </template>

        <!-- Date: formatted -->
        <template v-else-if="columnTypes[col.key] === 'date'">
          <HDateTime :value="value" />
        </template>

        <!-- Default: plain text -->
        <template v-else>
          {{ value }}
        </template>
      </template>
    </HDataTable>

    <!-- Detail Drawer -->
    <HDetailDrawer
      v-if="drawerRow"
      :title="drawerRow.name ?? drawerRow.id ?? title"
      :fields="drawerFields(drawerRow)"
      @close="drawerRow = null"
    />

    <!-- JSON Viewer Modal -->
    <HJsonViewer
      v-if="jsonViewerData !== null"
      :data="jsonViewerData"
      :title="jsonViewerTitle"
      @close="jsonViewerData = null"
    />
  </div>
</template>
