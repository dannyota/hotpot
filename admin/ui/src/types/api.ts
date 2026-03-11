export interface FilterOption {
  value: string
  count: number
  label?: string
}

/** Standard paginated list response from the API. */
export interface ListResponse<T> {
  data: T[]
  meta: PaginationMeta
  filter_options?: Record<string, FilterOption[]>
}

export interface PaginationMeta {
  page: number
  size: number
  total: number
  total_pages: number
}

export interface ErrorResponse {
  error: { code: number; message: string }
}

/** Query parameters for list endpoints. */
export interface ListParams {
  page?: number
  size?: number
  sort?: string
  filters?: Record<string, string>
}

// ─── Silver: Inventory Machine ──────────────────────────
export interface InventoryMachine {
  id: string
  hostname: string
  os_type: string
  os_name: string
  status: string
  internal_ip: string
  external_ip: string
  environment: string
  cloud_provider: string
  cloud_project: string
  cloud_zone: string
  collected_at: string
}

// ─── Gold: Lifecycle Software ───────────────────────────
export interface LifecycleSoftware {
  id: string
  name: string
  version: string
  classification: string
  eol_status: string
  eol_date: string
  eoes_date: string
  collected_at: string
}

// ─── Stats ──────────────────────────────────────────────
export interface StatValue {
  count: number
  delta?: number
}

export interface BronzeResource {
  label: string
  count: number
  delta?: number
}

export interface BronzeProvider {
  resources: BronzeResource[]
}

export interface StatsOverview {
  bronze: Record<string, BronzeProvider>
  silver: Record<string, StatValue>
  gold: Record<string, StatValue>
}
