<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ChevronRight, Star } from 'lucide-vue-next'
import { useUIConfig, type NavItem } from '@/composables/useUIConfig'
import { useFavorites } from '@/composables/useFavorites'

const route = useRoute()
const router = useRouter()
const { config } = useUIConfig()
const { favorites, isFavorite, toggleFavorite } = useFavorites()

const FAVORITES_LABEL = '★ Favorites'

const breadcrumbs = computed<string[]>(() => {
  return (route.meta?.breadcrumb as string[]) ?? []
})

const currentPath = computed(() => route.path)
const starred = computed(() => isFavorite(currentPath.value))

// --- Cascading dropdown state ---

const open = ref(false)
const selection = ref<string[]>([])

let showTimer: ReturnType<typeof setTimeout> | null = null
let hideTimer: ReturnType<typeof setTimeout> | null = null

function clearTimers() {
  if (showTimer) { clearTimeout(showTimer); showTimer = null }
  if (hideTimer) { clearTimeout(hideTimer); hideTimer = null }
}

onBeforeUnmount(clearTimers)

// Close dropdown on route change (e.g. sidebar click while open).
watch(() => route.path, () => {
  clearTimers()
  open.value = false
})

function onEnter() {
  clearTimers()
  showTimer = setTimeout(() => {
    selection.value = [...breadcrumbs.value]
    open.value = true
  }, 300)
}

function onLeave() {
  clearTimers()
  hideTimer = setTimeout(() => { open.value = false }, 150)
}

function cancelHide() {
  if (hideTimer) { clearTimeout(hideTimer); hideTimer = null }
}

// --- Column builder ---

const firstColumn = computed<NavItem[]>(() => {
  const items = [...config.value.nav]
  if (favorites.value.length) {
    const favChildren: NavItem[] = favorites.value.map(f => ({
      label: f.context ? `${f.label} (${f.context})` : f.label,
      path: f.path,
    }))
    items.unshift({ label: FAVORITES_LABEL, children: favChildren })
  }
  return items
})

const columns = computed(() => {
  const cols: NavItem[][] = []
  const topLevel = firstColumn.value
  if (!topLevel.length) return cols
  cols.push(topLevel)

  let current: NavItem[] = topLevel
  for (const sel of selection.value) {
    const match = current.find(item => item.label === sel)
    if (!match?.children?.length) break
    cols.push(match.children)
    current = match.children
  }
  return cols
})

function select(depth: number, label: string) {
  selection.value = [...selection.value.slice(0, depth), label]
}

function navigate(path: string) {
  router.push(path)
  open.value = false
}
</script>

<template>
  <div
    class="relative flex items-center gap-1.5 text-sm text-zinc-500 dark:text-zinc-400 shrink-0"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
  >
    <template v-for="(b, i) in breadcrumbs" :key="i">
      <span v-if="i > 0" class="text-zinc-300 dark:text-zinc-600">/</span>
      <span :class="i === breadcrumbs.length - 1 ? 'text-zinc-900 dark:text-zinc-100 font-medium' : ''">
        {{ b }}
      </span>
    </template>
    <button
      v-if="breadcrumbs.length"
      class="p-1 rounded transition-colors"
      :class="starred
        ? 'text-amber-400'
        : 'text-zinc-300 dark:text-zinc-600 hover:text-amber-400'"
      @click="toggleFavorite(currentPath)"
    >
      <Star class="w-3.5 h-3.5" :fill="starred ? 'currentColor' : 'none'" />
    </button>

    <!-- Cascading column picker -->
    <div
      v-if="open"
      class="absolute top-full left-0 pt-1 z-50"
      @mouseenter="cancelHide"
    >
      <div class="flex rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-lg">
        <div
          v-for="(col, depth) in columns"
          :key="depth"
          class="min-w-[150px] max-h-[60vh] overflow-y-auto py-1"
          :class="depth < columns.length - 1 && 'border-r border-zinc-200 dark:border-zinc-700'"
        >
          <button
            v-for="item in col"
            :key="item.label"
            class="flex items-center gap-1 w-full px-3 py-1.5 text-sm text-left transition-colors"
            :class="selection[depth] === item.label
              ? 'text-zinc-900 dark:text-zinc-100 bg-zinc-100 dark:bg-zinc-800 font-medium'
              : 'text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100 hover:bg-zinc-50 dark:hover:bg-zinc-800'"
            @mouseenter="select(depth, item.label)"
            @mousedown.prevent="item.path ? navigate(item.path) : undefined"
          >
            <span class="flex-1 truncate">{{ item.label }}</span>
            <ChevronRight v-if="item.children" class="w-3 h-3 shrink-0 text-zinc-400" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
