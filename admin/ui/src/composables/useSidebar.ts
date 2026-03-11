import { ref, computed, watch } from 'vue'

const pinned = ref(localStorage.getItem('sidebar-collapsed') !== 'true')
const tempOpen = ref(false)
const expanded = ref<Set<string>>(loadExpanded())

function loadExpanded(): Set<string> {
  try {
    const raw = localStorage.getItem('sidebar-expanded')
    if (raw) return new Set(JSON.parse(raw))
  } catch { /* ignore */ }
  return new Set()
}

// Auto-save whenever expanded changes (covers both toggleGroup and NavTreeNode.toggle)
watch(expanded, (s) => {
  localStorage.setItem('sidebar-expanded', JSON.stringify([...s]))
})

export function useSidebar() {
  const open = computed(() => pinned.value || tempOpen.value)

  function togglePin() {
    pinned.value = !pinned.value
    tempOpen.value = false
    localStorage.setItem('sidebar-collapsed', String(!pinned.value))
  }

  function toggleGroup(label: string) {
    if (!pinned.value) {
      // Collapsed — temp-expand and show this group.
      const s = new Set(expanded.value)
      s.add(label)
      expanded.value = s
      tempOpen.value = true
      return
    }
    const s = new Set(expanded.value)
    if (s.has(label)) s.delete(label)
    else s.add(label)
    expanded.value = s
  }

  function closeTemp() {
    tempOpen.value = false
  }

  return { open, pinned, expanded, togglePin, toggleGroup, closeTemp }
}
