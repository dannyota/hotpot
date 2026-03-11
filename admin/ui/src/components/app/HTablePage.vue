<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useTableQuery } from '@/composables/useTableQuery'
import { exportCsv } from '@/composables/useExport'
import { useColumnChooser } from '@/composables/useColumnChooser'
import { useTimezone } from '@/composables/useTimezone'
import { columnLabel } from '@/composables/columns'
import type { Column } from '@/components/app/HDataTable.vue'
import type { ColumnDef, DrawerFieldDef, FilterDef } from '@/types/table'
import HDataTable from '@/components/app/HDataTable.vue'
import HCellRenderer from '@/components/app/HCellRenderer.vue'
import HMultiSelect from '@/components/app/HMultiSelect.vue'
import HColumnChooser from '@/components/app/HColumnChooser.vue'
import HRefreshButton from '@/components/app/HRefreshButton.vue'
import HDetailDrawer from '@/components/app/HDetailDrawer.vue'
import HJsonViewer from '@/components/app/HJsonViewer.vue'
import { Search, X, Download, Eye } from 'lucide-vue-next'

const { formatDateTime } = useTimezone()
const router = useRouter()

const props = withDefaults(defineProps<{
  endpoint: string
  columns: ColumnDef[]
  drawerFields?: DrawerFieldDef[]
  filters?: FilterDef[]
  defaultSort?: string
  searchKey?: string
  searchPlaceholder?: string
  exportFilename?: string
  drawerTitleKey?: string
  title?: string
  stats?: boolean
  detailRoute?: (row: Record<string, any>) => string
}>(), {
  searchKey: 'name',
  searchPlaceholder: 'Search by name...',
  drawerTitleKey: 'name',
  title: 'Data',
  defaultSort: '',
})

// --- Table Query ---
const {
  data, meta, filterOptions, loading, error, sort, filters: queryFilters,
  reload, setSort, setPage, setPageSize,
  searchText, multiFilters, onSearchInput, onMultiFilterChange, hasActiveFilters, onClearAll,
} = useTableQuery<Record<string, any>>({
  endpoint: props.endpoint,
  defaultSort: props.defaultSort,
  searchKey: props.searchKey,
})

// --- Columns → HDataTable format ---
const tableColumns = computed<Column<Record<string, any>>[]>(() => {
  const cols = props.columns.map(c => ({
    key: c.key,
    label: c.label ?? columnLabel(c.key),
    sortable: c.sortable ?? true,
    ...(c.maxWidth != null && { maxWidth: c.maxWidth }),
  }))
  if (props.detailRoute) {
    cols.unshift({ key: '_detail', label: '', sortable: false })
  }
  return cols
})

const { visibleColumns, hiddenKeys, setVisible, resetAll } = useColumnChooser(props.endpoint, tableColumns)

// --- Filter helpers ---
function filterLabel(f: FilterDef): string {
  return f.label ?? f.key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
}

function getFilterOptions(f: FilterDef) {
  const opts = filterOptions.value[f.key] ?? []
  if (!f.bool) return opts
  return opts.map(o => ({
    ...o,
    value: o.value === 'true' ? 'Yes' : o.value === 'false' ? 'No' : o.value,
  }))
}

function getFilterValue(f: FilterDef): string[] {
  const vals = multiFilters.value[f.key] ?? []
  if (!f.bool) return vals
  return vals.map(v => v === 'true' ? 'Yes' : v === 'false' ? 'No' : v)
}

function onFilterChange(f: FilterDef, values: string[]) {
  const mapped = f.bool
    ? values.map(v => v === 'Yes' ? 'true' : v === 'No' ? 'false' : v)
    : values
  onMultiFilterChange(f.key, mapped)
}

// --- Stats ---
const statsData = ref<any>({})

async function fetchStats() {
  if (!props.stats) return
  try {
    const params = new URLSearchParams()
    for (const [k, v] of Object.entries(queryFilters.value)) {
      if (v) params.set(`filter[${k}]`, v)
    }
    const qs = params.toString()
    const res = await window.fetch(`${props.endpoint}/stats${qs ? '?' + qs : ''}`)
    if (res.ok) statsData.value = await res.json()
  } catch { /* ignore */ }
}

// --- Drawer ---
const drawerRow = ref<Record<string, any> | null>(null)

function resolveDrawerFields(row: Record<string, any>) {
  if (!props.drawerFields) {
    return Object.entries(row)
      .filter(([k]) => !k.startsWith('edges'))
      .map(([k, v]) => ({
        label: columnLabel(k),
        value: v,
        mono: k.endsWith('_id') || k === 'id',
      }))
  }
  return props.drawerFields.map(def => {
    const raw = row[def.key]
    let value: any = raw
    if (def.transform) {
      value = def.transform(raw, row)
    } else if (def.format === 'bool') {
      value = raw != null ? (raw ? 'Yes' : 'No') : null
    } else if (def.format === 'date') {
      value = raw ? formatDateTime(raw) : null
    }
    return { label: def.label, value, mono: def.mono }
  })
}

// --- Export ---
const exporting = ref(false)

async function onExport() {
  exporting.value = true
  try {
    await exportCsv(props.endpoint, queryFilters.value, sort.value, meta.value.total, props.exportFilename ?? 'export.csv')
  } finally {
    exporting.value = false
  }
}

// --- JSON Viewer ---
const jsonViewerData = ref<any>(null)
const jsonViewerTitle = ref('')

function showJson(data: any, title: string) {
  jsonViewerData.value = data
  jsonViewerTitle.value = title
}

// --- Emit ---
const emit = defineEmits<{
  data: [rows: Record<string, any>[]]
}>()

// --- Lifecycle ---
watch(data, (rows) => emit('data', rows))
watch(queryFilters, () => fetchStats(), { deep: true })
onMounted(() => {
  reload()
  if (props.stats) fetchStats()
})
</script>

<template>
  <div class="p-6 space-y-2 max-w-full">
    <!-- Header: title/stats + actions -->
    <div class="flex items-center justify-between gap-4">
      <slot name="title" :stats="statsData" :meta="meta" :loading="loading">
        <h1 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">{{ title }}</h1>
      </slot>
      <div class="flex items-center gap-2">
        <HRefreshButton :loading="loading" />
        <HColumnChooser :columns="tableColumns" :hidden-keys="hiddenKeys" @toggle="setVisible" @reset="resetAll" />
        <button
          v-if="meta.total > 0 && exportFilename"
          :disabled="exporting"
          class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-zinc-700 dark:text-zinc-300 border border-zinc-300 dark:border-zinc-600 rounded-md hover:bg-zinc-50 dark:hover:bg-zinc-800 disabled:opacity-50 transition-colors shrink-0"
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
          :placeholder="searchPlaceholder"
          class="w-full pl-8 pr-3 py-1.5 text-sm border border-zinc-200 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-900 text-zinc-900 dark:text-zinc-100 placeholder:text-zinc-400 focus:outline-none focus:ring-2 focus:ring-zinc-900/10 dark:focus:ring-zinc-100/10 transition-shadow"
          @input="onSearchInput"
        />
      </div>
      <HMultiSelect
        v-for="f in filters"
        :key="f.key"
        :label="filterLabel(f)"
        :options="getFilterOptions(f)"
        :model-value="getFilterValue(f)"
        @update:model-value="onFilterChange(f, $event)"
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
      <template v-if="props.detailRoute" #_detail="{ row }">
        <button
          class="p-1 rounded hover:bg-zinc-200 dark:hover:bg-zinc-700 text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-200 transition-colors"
          title="View detail"
          @click.stop="router.push(props.detailRoute!(row))"
        >
          <Eye class="w-4 h-4" />
        </button>
      </template>
      <template v-for="col in columns" :key="col.key" #[col.key]="{ value, row }">
        <slot :name="`cell-${col.key}`" :value="value" :row="row" :show-json="showJson">
          <HCellRenderer :col="col" :value="value" :row="row" @show-json="showJson" />
        </slot>
      </template>
    </HDataTable>

    <!-- Drawer -->
    <HDetailDrawer
      v-if="drawerRow"
      :title="drawerRow[drawerTitleKey]"
      :fields="resolveDrawerFields(drawerRow)"
      @close="drawerRow = null"
    />

    <!-- JSON Viewer -->
    <HJsonViewer
      v-if="jsonViewerData !== null"
      :data="jsonViewerData"
      :title="jsonViewerTitle"
      @close="jsonViewerData = null"
    />
  </div>
</template>
