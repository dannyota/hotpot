import { ref, computed } from 'vue'

export type NotificationType = 'error' | 'warning' | 'info'

export interface Notification {
  id: string
  type: NotificationType
  message: string
  source: string
  timestamp: number
  read: boolean
}

export type Toast = Pick<Notification, 'id' | 'type' | 'message' | 'source'>

const STORAGE_KEY = 'hotpot-notifications'
const MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000 // 30 days
const TOAST_DURATION = 5000
const DEDUPE_WINDOW = 5 * 60 * 1000 // 5 minutes

function load(): Notification[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const items: Notification[] = JSON.parse(raw)
    const cutoff = Date.now() - MAX_AGE_MS
    return items.filter(n => n.timestamp > cutoff)
  } catch {
    return []
  }
}

function save(items: Notification[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(items))
}

const notifications = ref<Notification[]>(load())
const toasts = ref<Toast[]>([])

const unreadCount = computed(() => notifications.value.filter(n => !n.read).length)

function add(type: NotificationType, message: string, source: string) {
  const now = Date.now()
  const dup = notifications.value.find(
    n => !n.read && n.source === source && n.message === message && now - n.timestamp < DEDUPE_WINDOW,
  )
  if (dup) return

  const id = crypto.randomUUID()

  notifications.value.unshift({ id, type, message, source, timestamp: now, read: false })
  save(notifications.value)

  toasts.value.push({ id, type, message, source })
  setTimeout(() => dismissToast(id), TOAST_DURATION)
}

function dismissToast(id: string) {
  toasts.value = toasts.value.filter(t => t.id !== id)
}

function markRead(id: string) {
  const n = notifications.value.find(n => n.id === id)
  if (n && !n.read) {
    n.read = true
    save(notifications.value)
  }
}

function markAllRead() {
  let changed = false
  for (const n of notifications.value) {
    if (!n.read) {
      n.read = true
      changed = true
    }
  }
  if (changed) save(notifications.value)
}

function remove(id: string) {
  notifications.value = notifications.value.filter(n => n.id !== id)
  save(notifications.value)
}

function clearAll() {
  notifications.value = []
  save(notifications.value)
}

/** Shorten an API endpoint for display. */
export function shortenSource(source: string): string {
  return source.replace(/^\/api\/v1\//, '').replace(/\?.*$/, '')
}

/** Format an epoch timestamp as relative time. */
export function formatNotificationTime(ts: number): string {
  const diff = Date.now() - ts
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

export function useNotifications() {
  return { notifications, toasts, unreadCount, add, dismissToast, markRead, markAllRead, remove, clearAll }
}
