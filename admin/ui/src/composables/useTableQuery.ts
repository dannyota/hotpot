import { ref, computed, watch, provide, type Ref } from 'vue'
import type { ListParams } from '@/types/api'
import { useListApi } from './useApi'

const STORAGE_PREFIX = 'hotpot-query:'

interface SavedQuery {
  filters?: Record<string, string>
  sort?: string
  size?: number
}

function loadSaved(key: string): SavedQuery {
  try {
    const raw = localStorage.getItem(STORAGE_PREFIX + key)
    if (raw) return JSON.parse(raw)
  } catch { /* ignore */ }
  return {}
}

function persistSaved(key: string, q: SavedQuery) {
  const clean: SavedQuery = {}
  const activeFilters = Object.fromEntries(
    Object.entries(q.filters ?? {}).filter(([, v]) => v !== ''),
  )
  if (Object.keys(activeFilters).length) clean.filters = activeFilters
  if (q.sort) clean.sort = q.sort
  if (q.size && q.size !== 20) clean.size = q.size

  if (Object.keys(clean).length === 0) {
    localStorage.removeItem(STORAGE_PREFIX + key)
  } else {
    localStorage.setItem(STORAGE_PREFIX + key, JSON.stringify(clean))
  }
}

function filtersToMulti(filters: Record<string, string>, searchKey: string): Record<string, string[]> {
  const mf: Record<string, string[]> = {}
  for (const [k, v] of Object.entries(filters)) {
    if (k !== searchKey && v) mf[k] = v.split(',')
  }
  return mf
}

export interface TableQueryOptions {
  endpoint: string | (() => string)
  defaultSort?: string
  defaultSize?: number
  /** Filter key used for the text search input (e.g. 'name', 'q', 'description'). Defaults to 'name'. */
  searchKey?: string
}

/**
 * Composable that wires pagination, sorting, filtering, and search to a list API.
 * Manages searchText and multiFilters internally so pages can't forget to restore them.
 * Persists filters, sort, and page size per endpoint in localStorage.
 */
export function useTableQuery<T>(opts: TableQueryOptions) {
  const api = useListApi<T>(opts.endpoint)
  const searchKey = opts.searchKey ?? 'name'

  const resolveKey = () => typeof opts.endpoint === 'function' ? opts.endpoint() : opts.endpoint

  // Load saved state for initial render.
  const saved = loadSaved(resolveKey())

  const page = ref(1)
  const size = ref(saved.size ?? opts.defaultSize ?? 20)
  const sort = ref(saved.sort ?? opts.defaultSort ?? '')
  const filters: Ref<Record<string, string>> = ref(saved.filters ?? {})

  // Search and multi-filter UI state — initialized from restored filters.
  const searchText = ref(filters.value[searchKey] ?? '')
  const multiFilters = ref<Record<string, string[]>>(filtersToMulti(filters.value, searchKey))

  let searchTimeout: ReturnType<typeof setTimeout> | null = null

  function onSearchInput() {
    if (searchTimeout) clearTimeout(searchTimeout)
    searchTimeout = setTimeout(() => {
      setFilter(searchKey, searchText.value)
    }, 300)
  }

  function onMultiFilterChange(field: string, values: string[]) {
    multiFilters.value = { ...multiFilters.value, [field]: values }
    setFilter(field, values.join(','))
  }

  const hasActiveFilters = computed(() =>
    searchText.value !== '' ||
    Object.values(multiFilters.value).some(v => v.length > 0),
  )

  function onClearAll() {
    searchText.value = ''
    multiFilters.value = {}
    clearFilters()
  }

  function buildParams(): ListParams {
    return {
      page: page.value,
      size: size.value,
      sort: sort.value || undefined,
      filters: Object.fromEntries(
        Object.entries(filters.value).filter(([, v]) => v !== '')
      ),
    }
  }

  async function reload() {
    await api.fetch(buildParams())
  }

  function setSort(field: string) {
    if (sort.value === field) {
      sort.value = `-${field}`
    } else if (sort.value === `-${field}`) {
      sort.value = ''
    } else {
      sort.value = field
    }
    page.value = 1
  }

  function setFilter(field: string, value: string) {
    filters.value = { ...filters.value, [field]: value }
    page.value = 1
  }

  function clearFilters() {
    filters.value = {}
    page.value = 1
  }

  function setPage(p: number) {
    page.value = p
  }

  function setPageSize(s: number) {
    size.value = s
    page.value = 1
  }

  // Persist filters/sort/size on change.
  watch([sort, filters, size], () => {
    const key = resolveKey()
    if (key) persistSaved(key, { filters: filters.value, sort: sort.value, size: size.value })
  }, { deep: true })

  // Resolved endpoint ref — used for both state-restore and auto-reload.
  const endpointRef = computed(resolveKey)

  // For function endpoints (GenericTablePage): restore saved state on endpoint change.
  // Created BEFORE the auto-reload watch so state is updated before reload fires.
  if (typeof opts.endpoint === 'function') {
    watch(endpointRef, (newKey) => {
      if (!newKey) return
      const s = loadSaved(newKey)
      filters.value = s.filters ?? {}
      sort.value = s.sort ?? opts.defaultSort ?? ''
      size.value = s.size ?? opts.defaultSize ?? 20
      page.value = 1
      // Sync UI state from restored filters.
      searchText.value = filters.value[searchKey] ?? ''
      multiFilters.value = filtersToMulti(filters.value, searchKey)
    })
  }

  // Expose reload to child components (e.g. HDataTable refresh button).
  provide('hotpot-table-reload', reload)

  // Auto-reload when page/size/sort/filters or endpoint change.
  watch([page, size, sort, filters, endpointRef], reload, { deep: true })

  return {
    data: api.data,
    meta: api.meta,
    filterOptions: api.filterOptions,
    loading: api.loading,
    error: api.error,
    page,
    size,
    sort,
    filters,
    reload,
    setSort,
    setFilter,
    clearFilters,
    setPage,
    setPageSize,
    // Search & multi-filter UI state — managed here so pages can't forget to restore them.
    searchText,
    multiFilters,
    onSearchInput,
    onMultiFilterChange,
    hasActiveFilters,
    onClearAll,
  }
}
