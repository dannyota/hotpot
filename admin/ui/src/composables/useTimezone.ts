import { ref, computed } from 'vue'

const STORAGE_KEY = 'hotpot-timezone'

/** Well-known timezone presets. */
export const TIMEZONE_OPTIONS = [
  { value: 'local', label: 'Local' },
  { value: 'UTC', label: 'UTC' },
  { value: 'America/New_York', label: 'US Eastern' },
  { value: 'America/Chicago', label: 'US Central' },
  { value: 'America/Denver', label: 'US Mountain' },
  { value: 'America/Los_Angeles', label: 'US Pacific' },
  { value: 'Europe/London', label: 'London' },
  { value: 'Europe/Paris', label: 'Paris' },
  { value: 'Europe/Berlin', label: 'Berlin' },
  { value: 'Asia/Tokyo', label: 'Tokyo' },
  { value: 'Asia/Shanghai', label: 'Shanghai' },
  { value: 'Asia/Singapore', label: 'Singapore' },
  { value: 'Asia/Ho_Chi_Minh', label: 'Ho Chi Minh' },
  { value: 'Australia/Sydney', label: 'Sydney' },
] as const

const timezone = ref(localStorage.getItem(STORAGE_KEY) || 'local')

export function useTimezone() {
  function setTimezone(tz: string) {
    timezone.value = tz
    localStorage.setItem(STORAGE_KEY, tz)
  }

  const label = computed(() => {
    const opt = TIMEZONE_OPTIONS.find(o => o.value === timezone.value)
    return opt?.label ?? timezone.value
  })

  /** Format an ISO datetime string as absolute date+time in the selected timezone. */
  function formatDateTime(iso: string): string {
    if (!iso) return ''
    const d = new Date(iso)
    const opts: Intl.DateTimeFormatOptions = {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    }
    if (timezone.value !== 'local') {
      opts.timeZone = timezone.value
    }
    return d.toLocaleString('en-US', opts)
  }

  /** Format an ISO datetime string as relative time (e.g. "5m ago"). Timezone-independent. */
  function relativeTime(iso: string): string {
    if (!iso) return ''
    const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000)
    if (diff < 0) return 'just now'
    if (diff < 60) return `${diff}s ago`
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
    return `${Math.floor(diff / 86400)}d ago`
  }

  return { timezone, label, setTimezone, formatDateTime, relativeTime }
}
