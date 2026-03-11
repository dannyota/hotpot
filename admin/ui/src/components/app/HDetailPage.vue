<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, reactive, isRef } from 'vue'
import { useRouter } from 'vue-router'
import { buildUrl } from '@/composables/useApi'
import { columnLabel } from '@/composables/columns'
import { boolBadgeClass } from '@/composables/formatting'
import type { Column } from '@/components/app/HDataTable.vue'
import type { DetailFieldDef, DetailTabDef } from '@/types/table'
import HDataTable from '@/components/app/HDataTable.vue'
import HCellRenderer from '@/components/app/HCellRenderer.vue'
import HDateTime from '@/components/app/HDateTime.vue'
import HRelativeTime from '@/components/app/HRelativeTime.vue'
import HJsonViewer from '@/components/app/HJsonViewer.vue'
import HDetailDrawer from '@/components/app/HDetailDrawer.vue'
import { ArrowLeft, Loader2, Check, Copy, RefreshCw } from 'lucide-vue-next'

const router = useRouter()

const props = withDefaults(defineProps<{
  endpoint: string
  backRoute: string
  backLabel?: string
  titleKey?: string
  subtitleFields?: string[]
  fields: DetailFieldDef[]
  tabs: DetailTabDef[]
}>(), {
  titleKey: 'name',
  backLabel: 'Back',
})

// --- Detail fetch ---
const detail = ref<Record<string, any> | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
let detailAbort: AbortController | null = null

async function fetchDetail() {
  detailAbort?.abort()
  detailAbort = new AbortController()
  loading.value = true
  error.value = null
  try {
    const res = await window.fetch(props.endpoint, { signal: detailAbort.signal })
    if (!res.ok) {
      const body = await res.json().catch(() => null)
      throw new Error(body?.error?.message || `HTTP ${res.status}`)
    }
    const json = await res.json()
    detail.value = json.data
  } catch (e: any) {
    if (e.name === 'AbortError') return
    error.value = e.message
    detail.value = null
  } finally {
    loading.value = false
  }
}

onMounted(fetchDetail)
watch(() => props.endpoint, () => {
  // Reset all API tab state to prevent stale data flash
  for (const key of Object.keys(apiTabs)) {
    apiTabs[key] = { data: [], meta: { page: 1, size: 10, total: 0, total_pages: 0 }, sort: '', loading: false, error: null, fetched: false }
  }
  fetchDetail()
})

// --- Active tab (persisted per resource type) ---
const TAB_STORAGE_PREFIX = 'hotpot-detail-tab:'
const tabStorageKey = computed(() => {
  // Strip the ID to get the resource base path, e.g. "/api/v1/.../instances/123" → "/api/v1/.../instances"
  const parts = props.endpoint.split('/')
  parts.pop()
  return TAB_STORAGE_PREFIX + parts.join('/')
})

function loadSavedTab(): string {
  try {
    const saved = localStorage.getItem(tabStorageKey.value)
    if (saved && props.tabs.some(t => t.key === saved)) return saved
  } catch { /* ignore */ }
  return props.tabs[0]?.key ?? ''
}

const activeTab = ref(loadSavedTab())

// --- Field rendering ---
function subtitleFieldDef(key: string): DetailFieldDef | undefined {
  return props.fields.find(f => f.key === key)
}

function fieldValue(def: DetailFieldDef, row: Record<string, any>): any {
  const raw = row[def.key]
  if (def.transform) return def.transform(raw, row)
  if (def.format === 'bool') return raw != null ? (raw ? 'Yes' : 'No') : null
  return raw
}

// --- Tab helpers ---
function tabColumns(tab: DetailTabDef): Column<Record<string, any>>[] {
  return tab.columns.map(c => ({
    key: c.key,
    label: c.label ?? columnLabel(c.key),
    sortable: c.sortable ?? !tab.edgeKey,
  }))
}

function edgeData(tab: DetailTabDef): Record<string, any>[] {
  if (!detail.value || !tab.edgeKey) return []
  const parts = tab.edgeKey.split('.')
  let v: any = detail.value
  for (const p of parts) {
    v = v?.[p]
  }
  return Array.isArray(v) ? v : []
}

function tabCount(tab: DetailTabDef): number | null {
  if (tab.edgeKey && detail.value) return edgeData(tab).length
  if (tab.apiEndpoint && apiTabs[tab.key]?.meta.total > 0) return apiTabs[tab.key].meta.total
  return null
}

// --- Edge tab local pagination ---
const edgePagination = ref<Record<string, { page: number; size: number }>>({})

function getEdgePage(tabKey: string) {
  return edgePagination.value[tabKey] ?? { page: 1, size: 10 }
}

function edgeMeta(tab: DetailTabDef) {
  const all = edgeData(tab)
  const p = getEdgePage(tab.key)
  const total = all.length
  return { page: p.page, size: p.size, total, total_pages: Math.ceil(total / p.size) || 1 }
}

function edgePagedData(tab: DetailTabDef) {
  const all = edgeData(tab)
  const p = getEdgePage(tab.key)
  const start = (p.page - 1) * p.size
  return all.slice(start, start + p.size)
}

function setEdgePage(tabKey: string, page: number) {
  const p = getEdgePage(tabKey)
  edgePagination.value = { ...edgePagination.value, [tabKey]: { ...p, page } }
}

function setEdgePageSize(tabKey: string, size: number) {
  edgePagination.value = { ...edgePagination.value, [tabKey]: { page: 1, size } }
}

// --- API tabs (lightweight, no useTableQuery) ---
interface ApiTabState {
  data: Record<string, any>[]
  meta: { page: number; size: number; total: number; total_pages: number }
  sort: string
  loading: boolean
  error: string | null
  fetched: boolean
}

const apiTabs = reactive<Record<string, ApiTabState>>({})

function getApiTab(key: string): ApiTabState {
  if (!apiTabs[key]) {
    apiTabs[key] = { data: [], meta: { page: 1, size: 10, total: 0, total_pages: 0 }, sort: '', loading: false, error: null, fetched: false }
  }
  return apiTabs[key]
}

function resolveApiEndpoint(tab: DetailTabDef): string {
  if (!tab.apiEndpoint) return ''
  return isRef(tab.apiEndpoint) ? tab.apiEndpoint.value : tab.apiEndpoint
}

async function fetchApiTab(tab: DetailTabDef) {
  const st = getApiTab(tab.key)
  const ep = resolveApiEndpoint(tab)
  if (!ep) return
  st.loading = true
  st.error = null
  try {
    const url = buildUrl(ep, { page: st.meta.page, size: st.meta.size, sort: st.sort || undefined })
    const res = await window.fetch(url)
    if (!res.ok) {
      const body = await res.json().catch(() => null)
      throw new Error(body?.error?.message || `HTTP ${res.status}`)
    }
    const json = await res.json()
    st.data = json.data
    st.meta = json.meta
    st.fetched = true
  } catch (e: any) {
    st.error = e.message
    st.data = []
  } finally {
    st.loading = false
  }
}

function apiTabSetSort(tab: DetailTabDef, field: string) {
  const st = getApiTab(tab.key)
  if (st.sort === field) st.sort = `-${field}`
  else if (st.sort === `-${field}`) st.sort = ''
  else st.sort = field
  st.meta.page = 1
  fetchApiTab(tab)
}

function apiTabSetPage(tab: DetailTabDef, page: number) {
  getApiTab(tab.key).meta.page = page
  fetchApiTab(tab)
}

function apiTabSetPageSize(tab: DetailTabDef, size: number) {
  const st = getApiTab(tab.key)
  st.meta.page = 1
  st.meta.size = size
  fetchApiTab(tab)
}

// Prefetch counts for all API tabs, and fully load the active one.
watch(detail, (d) => {
  if (!d) return
  for (const tab of props.tabs) {
    if (!tab.apiEndpoint) continue
    const ep = resolveApiEndpoint(tab)
    if (!ep) continue
    // Active tab: full fetch. Others: count only.
    if (tab.key === activeTab.value) {
      fetchApiTab(tab)
    } else {
      const st = getApiTab(tab.key)
      window.fetch(buildUrl(ep, { page: 1, size: 1 }))
        .then(r => r.ok ? r.json() : null)
        .then(json => { if (json?.meta) { st.meta.total = json.meta.total; st.meta.total_pages = json.meta.total_pages } })
        .catch(() => {})
    }
  }
})

// Persist active tab + lazy-load API tab data on switch.
watch(activeTab, (key) => {
  try { localStorage.setItem(tabStorageKey.value, key) } catch { /* ignore */ }
  const tab = props.tabs.find(t => t.key === key)
  if (tab?.apiEndpoint && !getApiTab(tab.key).fetched) fetchApiTab(tab)
})

// --- Refresh ---
function refresh() {
  // Reset API tab state so they refetch.
  for (const key of Object.keys(apiTabs)) {
    apiTabs[key].fetched = false
  }
  fetchDetail()
}

// --- Copy to clipboard ---
const copiedKey = ref<string | null>(null)
let copyTimer: ReturnType<typeof setTimeout> | null = null

function copyField(key: string) {
  if (!detail.value) return
  const f = props.fields.find(fd => fd.key === key)
  const raw = detail.value[key]
  const text = f?.transform ? f.transform(raw, detail.value) : String(raw ?? '')
  navigator.clipboard.writeText(text)
  copiedKey.value = key
  if (copyTimer) clearTimeout(copyTimer)
  copyTimer = setTimeout(() => { copiedKey.value = null; copyTimer = null }, 1500)
}

// --- Subtable drawer ---
const drawerRow = ref<Record<string, any> | null>(null)

function drawerFields(row: Record<string, any>): { label: string; value: any; mono: boolean }[] {
  return Object.entries(row).map(([k, v]) => ({
    label: columnLabel(k),
    value: v,
    mono: k.endsWith('_id') || k === 'id',
  }))
}

// --- JSON Viewer ---
const jsonViewerData = ref<any>(null)
const jsonViewerTitle = ref('')

function showJson(data: any, title: string) {
  jsonViewerData.value = data
  jsonViewerTitle.value = title
}

onUnmounted(() => {
  detailAbort?.abort()
  if (copyTimer) clearTimeout(copyTimer)
})
</script>

<template>
  <div class="p-6 space-y-4 max-w-full">
    <!-- Back link + refresh -->
    <div class="flex items-center justify-between">
      <button
        class="inline-flex items-center gap-1.5 text-sm text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors"
        @click="router.push(backRoute)"
      >
        <ArrowLeft class="w-4 h-4" />
        {{ backLabel }}
      </button>
      <button
        v-if="detail"
        :disabled="loading"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-zinc-700 dark:text-zinc-300 border border-zinc-300 dark:border-zinc-600 rounded-md hover:bg-zinc-50 dark:hover:bg-zinc-800 disabled:opacity-50 transition-colors shrink-0"
        @click="refresh"
      >
        <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" />
        Refresh
      </button>
    </div>

    <!-- Loading / Error -->
    <div v-if="loading && !detail" class="flex items-center justify-center py-20">
      <Loader2 class="w-6 h-6 animate-spin text-zinc-400" />
    </div>

    <div v-else-if="error" class="rounded-lg border border-red-200 dark:border-red-900 bg-red-50 dark:bg-red-950 p-4">
      <p class="text-sm text-red-700 dark:text-red-400">{{ error }}</p>
    </div>

    <template v-else-if="detail">
      <!-- Title + subtitle chips -->
      <div class="space-y-1.5">
        <h1 class="text-xl font-semibold text-zinc-900 dark:text-zinc-100">{{ detail[titleKey] }}</h1>
        <div v-if="subtitleFields?.length" class="flex items-center gap-2 flex-wrap">
          <template v-for="sf in subtitleFields" :key="sf">
            <template v-if="detail[sf] != null && detail[sf] !== ''">
              <!-- Badge field: render with its own color, no gray wrapper -->
              <span
                v-if="subtitleFieldDef(sf)?.badge"
                class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full"
                :class="subtitleFieldDef(sf)!.badge!(detail[sf])"
              >{{ subtitleFieldDef(sf)?.transform?.(detail[sf], detail) ?? String(detail[sf]).replace(/_/g, ' ') }}</span>
              <!-- Plain field: gray chip -->
              <span
                v-else
                class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300"
              >{{ subtitleFieldDef(sf)?.transform?.(detail[sf], detail) ?? detail[sf] }}</span>
            </template>
          </template>
        </div>
      </div>

      <!-- Overview grid -->
      <div class="rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 p-4">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-3">
          <div v-for="f in fields" :key="f.key" class="group/field flex flex-col gap-0.5 min-w-0" :class="f.fullWidth ? 'sm:col-span-2 lg:col-span-3' : ''">
            <dt class="text-xs font-medium text-zinc-500 dark:text-zinc-400">{{ f.label }}</dt>
            <dd class="min-w-0 flex items-start gap-1">
              <div class="min-w-0 flex-1">
                <template v-if="f.format === 'date'">
                  <HDateTime :value="detail[f.key]" />
                </template>
                <template v-else-if="f.format === 'relative'">
                  <HRelativeTime :value="detail[f.key]" />
                </template>
                <template v-else-if="f.badge">
                  <span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full" :class="f.badge(detail[f.key])">
                    {{ fieldValue(f, detail) }}
                  </span>
                </template>
                <template v-else-if="f.format === 'bool'">
                  <span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full" :class="boolBadgeClass(detail[f.key])">
                    {{ detail[f.key] ? 'Yes' : 'No' }}
                  </span>
                </template>
                <span
                  v-else-if="f.mono"
                  class="block font-mono text-xs text-zinc-500 dark:text-zinc-400 truncate"
                  :title="String(fieldValue(f, detail) ?? '')"
                >
                  {{ fieldValue(f, detail) ?? '\u2014' }}
                </span>
                <span v-else class="text-sm text-zinc-900 dark:text-zinc-100">
                  {{ fieldValue(f, detail) ?? '\u2014' }}
                </span>
              </div>
              <button
                v-if="detail[f.key] != null && detail[f.key] !== ''"
                class="shrink-0 p-0.5 rounded text-zinc-300 dark:text-zinc-600 opacity-0 group-hover/field:opacity-100 hover:!text-zinc-500 dark:hover:!text-zinc-400 transition-all"
                title="Copy"
                @click="copyField(f.key)"
              >
                <Check v-if="copiedKey === f.key" class="w-3.5 h-3.5 text-emerald-500" />
                <Copy v-else class="w-3.5 h-3.5" />
              </button>
            </dd>
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div v-if="tabs.length > 0">
        <div class="border-b border-zinc-200 dark:border-zinc-800">
          <nav class="flex gap-0 -mb-px overflow-x-auto">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              class="px-4 py-2 text-sm font-medium border-b-2 transition-colors whitespace-nowrap"
              :class="activeTab === tab.key
                ? 'border-zinc-900 dark:border-zinc-100 text-zinc-900 dark:text-zinc-100'
                : 'border-transparent text-zinc-500 dark:text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-300 hover:border-zinc-300 dark:hover:border-zinc-600'"
              @click="activeTab = tab.key"
            >
              {{ tab.label }}
              <span
                v-if="tabCount(tab) != null"
                class="ml-1.5 inline-flex items-center justify-center px-1.5 py-0.5 text-xs rounded-full bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400"
              >{{ tabCount(tab) }}</span>
            </button>
          </nav>
        </div>

        <!-- Tab content -->
        <div class="mt-2">
          <template v-for="tab in tabs" :key="tab.key">
            <!-- Edge tab -->
            <div v-if="activeTab === tab.key && tab.edgeKey">
              <HDataTable
                :columns="tabColumns(tab)"
                :data="edgePagedData(tab)"
                :meta="edgeMeta(tab)"
                :loading="false"
                @page="(p) => setEdgePage(tab.key, p)"
                @page-size="(s) => setEdgePageSize(tab.key, s)"
                @row-click="(row) => drawerRow = row"
              >
                <template v-for="col in tab.columns" :key="col.key" #[col.key]="{ value, row }">
                  <HCellRenderer :col="col" :value="value" :row="row" @show-json="showJson" />
                </template>
              </HDataTable>
            </div>

            <!-- API tab -->
            <div v-else-if="activeTab === tab.key && tab.apiEndpoint">
              <HDataTable
                :columns="tabColumns(tab)"
                :data="getApiTab(tab.key).data"
                :meta="getApiTab(tab.key).meta"
                :sort="getApiTab(tab.key).sort"
                :loading="getApiTab(tab.key).loading"
                :error="getApiTab(tab.key).error"
                @sort="(f) => apiTabSetSort(tab, f)"
                @page="(p) => apiTabSetPage(tab, p)"
                @page-size="(s) => apiTabSetPageSize(tab, s)"
                @row-click="(row) => drawerRow = row"
              >
                <template v-for="col in tab.columns" :key="col.key" #[col.key]="{ value, row }">
                  <HCellRenderer :col="col" :value="value" :row="row" @show-json="showJson" />
                </template>
              </HDataTable>
            </div>
          </template>
        </div>
      </div>
    </template>

    <!-- Subtable row drawer -->
    <HDetailDrawer
      v-if="drawerRow"
      :fields="drawerFields(drawerRow)"
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
