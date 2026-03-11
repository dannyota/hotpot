/** Shared column label overrides — used by both dedicated and generic pages. */
const labelOverrides: Record<string, string> = {
  first_collected_at: 'First Seen',
  collected_at: 'Last Seen',
  normalized_at: 'Normalized',
}

/** Returns a human-friendly label for a column key. */
export function columnLabel(key: string): string {
  return labelOverrides[key] ?? key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
}
