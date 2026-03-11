import { ref, type Ref } from 'vue'
import type { ListResponse, ListParams, FilterOption } from '@/types/api'
import { useNotifications } from '@/composables/useNotifications'

const { add: addNotification } = useNotifications()

/** Build a URL with query parameters for list endpoints. */
export function buildUrl(base: string, params: ListParams): string {
  const u = new URL(base, window.location.origin)
  if (params.page) u.searchParams.set('page', String(params.page))
  if (params.size) u.searchParams.set('size', String(params.size))
  if (params.sort) u.searchParams.set('sort', params.sort)
  if (params.filters) {
    for (const [k, v] of Object.entries(params.filters)) {
      if (v) u.searchParams.set(`filter[${k}]`, v)
    }
  }
  return u.pathname + u.search
}

/** Composable for fetching a paginated list from the API. */
export function useListApi<T>(endpoint: string | (() => string)) {
  const data: Ref<T[]> = ref([])
  const meta = ref({ page: 1, size: 20, total: 0, total_pages: 0 })
  const filterOptions: Ref<Record<string, FilterOption[]>> = ref({})
  const loading = ref(false)
  const error: Ref<string | null> = ref(null)

  async function fetch(params: ListParams = {}) {
    const ep = typeof endpoint === 'function' ? endpoint() : endpoint
    loading.value = true
    error.value = null
    try {
      const url = buildUrl(ep, params)
      const res = await window.fetch(url)
      if (!res.ok) {
        const body = await res.json().catch(() => null)
        throw new Error(body?.error?.message || `HTTP ${res.status}`)
      }
      const json: ListResponse<T> = await res.json()
      data.value = json.data
      meta.value = json.meta
      if (json.filter_options) {
        filterOptions.value = json.filter_options
      }
    } catch (e: any) {
      error.value = e.message
      data.value = []
      addNotification('error', e.message, ep)
    } finally {
      loading.value = false
    }
  }

  return { data, meta, filterOptions, loading, error, fetch }
}

/** Composable for fetching a single JSON object from the API. */
export function useApi<T>(endpoint: string) {
  const data: Ref<T | null> = ref(null)
  const loading = ref(false)
  const error: Ref<string | null> = ref(null)

  async function fetch() {
    loading.value = true
    error.value = null
    try {
      const res = await window.fetch(endpoint)
      if (!res.ok) {
        const body = await res.json().catch(() => null)
        throw new Error(body?.error?.message || `HTTP ${res.status}`)
      }
      const json = await res.json()
      data.value = json.data
    } catch (e: any) {
      error.value = e.message
      data.value = null
      addNotification('error', e.message, endpoint)
    } finally {
      loading.value = false
    }
  }

  return { data, loading, error, fetch }
}
