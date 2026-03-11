export interface ColumnDef {
  key: string
  label?: string
  sortable?: boolean
  format?: 'bold' | 'mono' | 'date' | 'relative' | 'bool' | 'json' | 'number'
  transform?: (value: any, row: any) => string
  badge?: (value: any) => string
  /** Max column width in pixels. Truncates with ellipsis and shows full value on hover. */
  maxWidth?: number
}

export interface DrawerFieldDef {
  key: string
  label: string
  mono?: boolean
  format?: 'date' | 'bool'
  transform?: (value: any, row: any) => string
}

export interface FilterDef {
  key: string
  label?: string
  bool?: boolean
}

export interface DetailFieldDef {
  key: string
  label: string
  mono?: boolean
  format?: 'date' | 'bool' | 'relative'
  transform?: (value: any, row: any) => string
  badge?: (value: any) => string
  /** Span full width of the overview grid (for long values like URLs). */
  fullWidth?: boolean
}

export interface DetailTabDef {
  key: string
  label: string
  columns: ColumnDef[]
  /** Edge tab: data from detail response edges (no API call). */
  edgeKey?: string
  /** API tab: fetches from endpoint (paginated). */
  apiEndpoint?: string | import('vue').Ref<string>
}
