/** Extract last path segment: "projects/foo/zones/us-east1-b" → "us-east1-b" */
export function shortPath(v: string): string {
  if (!v) return ''
  return v.split('/').pop()!
}

/** Standard badge color palette */
export const badge = {
  emerald: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400',
  zinc: 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400',
  blue: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  amber: 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400',
  red: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  purple: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
} as const

/** Build a badge function from uppercase-value → color map */
export function badgeColors(
  map: Record<string, string>,
  fallback = badge.zinc,
): (v: string) => string {
  return (v: string) => map[v?.toUpperCase()] ?? fallback
}

/** Shorten GCP service account email: strip .iam.gserviceaccount.com domain */
export function shortEmail(v: string): string {
  if (!v) return ''
  return v.replace(/\.iam\.gserviceaccount\.com$/, '')
}

/** Replace underscores with spaces (e.g. "PRIVATE_SERVICE_CONNECT" → "PRIVATE SERVICE CONNECT") */
export function humanize(v: string): string {
  if (!v) return ''
  return v.replace(/_/g, ' ')
}

/** Format a numeric value with a unit suffix (e.g. 100, "GB" → "100 GB") */
export function formatUnit(value: number, unit: string): string {
  if (value == null) return ''
  return `${value} ${unit}`
}

/** Format byte count into human-readable string (e.g. 1073741824 → "1.0 GB") */
export function formatBytes(bytes: number): string {
  if (bytes == null) return ''
  if (bytes === 0) return '0 B'
  if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(1)} GB`
  if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(1)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${bytes} B`
}

// --- Column cell rendering helpers (shared by HTablePage, HDetailPage) ---

import type { ColumnDef } from '@/types/table'
import { columnLabel } from '@/composables/columns'

/** Resolve the display value for a column cell: apply transform or humanize badge. */
export function displayValue(col: ColumnDef, value: any, row: any): any {
  if (col.transform) return col.transform(value, row)
  if (col.badge && typeof value === 'string') return value.replace(/_/g, ' ')
  return value
}

/** Tailwind classes for a boolean badge (Yes = green, No = gray). */
export function boolBadgeClass(value: boolean): string {
  return value
    ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400'
    : 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400'
}

/** Standard dot colors for stats */
export const dot = {
  emerald: 'bg-emerald-500',
  emeraldPulse: 'bg-emerald-500 animate-pulse',
  zinc: 'bg-zinc-400',
  blue: 'bg-blue-500',
  bluePulse: 'bg-blue-500 animate-pulse',
  amber: 'bg-amber-500',
  amberPulse: 'bg-amber-500 animate-pulse',
  red: 'bg-red-500',
  redPulse: 'bg-red-500 animate-pulse',
} as const
