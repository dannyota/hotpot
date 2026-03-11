import type { ListResponse } from '@/types/api'
import { buildUrl } from '@/composables/useApi'

function toCsv(rows: Record<string, any>[]): string {
  if (rows.length === 0) return ''
  const keys = Object.keys(rows[0])
  const escape = (v: any): string => {
    if (v == null) return ''
    const s = typeof v === 'object' ? JSON.stringify(v) : String(v)
    if (s.includes(',') || s.includes('"') || s.includes('\n')) {
      return '"' + s.replace(/"/g, '""') + '"'
    }
    return s
  }
  const header = keys.join(',')
  const body = rows.map(row => keys.map(k => escape(row[k])).join(',')).join('\n')
  return header + '\n' + body
}

function download(csv: string, filename: string) {
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

/**
 * Export all filtered rows as CSV.
 * Fetches with the current filters but overrides size to get all rows.
 */
export async function exportCsv(
  endpoint: string,
  filters: Record<string, string>,
  sort: string,
  total: number,
  filename: string,
) {
  const size = Math.min(total || 10000, 10000)
  const url = buildUrl(endpoint, {
    page: 1,
    size,
    sort: sort || undefined,
    filters: Object.fromEntries(
      Object.entries(filters).filter(([, v]) => v !== '')
    ),
  })
  const res = await window.fetch(url)
  if (!res.ok) throw new Error(`Export failed: HTTP ${res.status}`)
  const json: ListResponse<Record<string, any>> = await res.json()
  const csv = toCsv(json.data)
  download(csv, filename)
  return json.data.length
}
