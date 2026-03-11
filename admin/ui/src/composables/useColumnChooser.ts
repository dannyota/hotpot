import { computed, isRef, ref, toValue, watch, type MaybeRef, type Ref } from 'vue'
import type { Column } from '@/components/app/HDataTable.vue'

const STORAGE_PREFIX = 'hotpot-columns:'

/**
 * Manages column visibility per page, persisted to localStorage.
 * Stores hidden column keys (not visible) so new columns auto-appear.
 *
 * pageKey can be a plain string (dedicated pages) or a reactive ref/computed
 * (GenericTablePage where the API endpoint changes on navigation).
 */
export function useColumnChooser(pageKey: MaybeRef<string>, allColumns: Ref<Column<any>[]>) {
  function storageKey() {
    return STORAGE_PREFIX + toValue(pageKey)
  }

  // Load hidden keys from localStorage for the current page key.
  function loadHidden(): Set<string> {
    try {
      const raw = localStorage.getItem(storageKey())
      if (raw) return new Set(JSON.parse(raw))
    } catch { /* ignore corrupt data */ }
    return new Set()
  }

  const hiddenKeys = ref<Set<string>>(loadHidden())

  function persist() {
    const key = storageKey()
    if (!key) return
    if (hiddenKeys.value.size === 0) {
      localStorage.removeItem(key)
    } else {
      localStorage.setItem(key, JSON.stringify([...hiddenKeys.value]))
    }
  }

  // When pageKey changes (GenericTablePage navigation), reload from localStorage.
  if (isRef(pageKey)) {
    watch(pageKey, () => {
      hiddenKeys.value = loadHidden()
    })
  }

  const visibleColumns = computed(() =>
    allColumns.value.filter(c => !hiddenKeys.value.has(c.key)),
  )

  function setVisible(key: string, visible: boolean) {
    const next = new Set(hiddenKeys.value)
    if (visible) next.delete(key)
    else next.add(key)
    hiddenKeys.value = next
    persist()
  }

  function resetAll() {
    hiddenKeys.value = new Set()
    persist()
  }

  // If allColumns changes (e.g. GenericTablePage loads new data), prune
  // hidden keys that no longer exist in the column set.
  watch(allColumns, (cols) => {
    const validKeys = new Set(cols.map(c => c.key))
    const pruned = new Set([...hiddenKeys.value].filter(k => validKeys.has(k)))
    if (pruned.size !== hiddenKeys.value.size) {
      hiddenKeys.value = pruned
      persist()
    }
  })

  return { visibleColumns, hiddenKeys, setVisible, resetAll }
}
