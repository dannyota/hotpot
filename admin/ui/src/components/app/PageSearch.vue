<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from 'lucide-vue-next'
import { useUIConfig, type NavItem } from '@/composables/useUIConfig'

const router = useRouter()
const { config } = useUIConfig()

const query = ref('')
const focused = ref(false)
const selectedIndex = ref(-1)

let blurTimer: ReturnType<typeof setTimeout> | null = null

onBeforeUnmount(() => {
  if (blurTimer) clearTimeout(blurTimer)
})

// --- Search index ---

interface SearchResult {
  label: string
  breadcrumb: string
  path: string
}

const allLeaves = computed(() => {
  const results: SearchResult[] = []
  function walk(items: NavItem[], trail: string[]) {
    for (const item of items) {
      if (item.children) {
        walk(item.children, [...trail, item.label])
      } else if (item.path) {
        results.push({ label: item.label, breadcrumb: trail.join(' > '), path: item.path })
      }
    }
  }
  walk(config.value.nav, [])
  return results
})

const results = computed(() => {
  const q = query.value.toLowerCase().trim()
  if (!q) return []
  const terms = q.split(/\s+/)
  return allLeaves.value.filter(leaf => {
    const text = `${leaf.breadcrumb} ${leaf.label}`.toLowerCase()
    return terms.every(t => text.includes(t))
  })
})

watch(results, () => { selectedIndex.value = -1 })

const showDropdown = computed(() => focused.value && query.value.trim().length > 0)

// --- Actions ---

function navigate(path: string) {
  router.push(path)
  query.value = ''
  focused.value = false
  selectedIndex.value = -1
}

function onKeydown(e: KeyboardEvent) {
  if (!showDropdown.value) return
  const len = results.value.length
  if (!len) return

  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      selectedIndex.value = (selectedIndex.value + 1) % len
      scrollToSelected()
      break
    case 'ArrowUp':
      e.preventDefault()
      selectedIndex.value = (selectedIndex.value - 1 + len) % len
      scrollToSelected()
      break
    case 'Enter':
      e.preventDefault()
      if (selectedIndex.value >= 0 && selectedIndex.value < len) {
        navigate(results.value[selectedIndex.value].path)
      }
      break
    case 'Escape':
      query.value = ''
      focused.value = false
      ;(e.target as HTMLInputElement)?.blur()
      break
  }
}

function scrollToSelected() {
  const el = document.querySelector('[data-search-selected="true"]')
  el?.scrollIntoView({ block: 'nearest' })
}

function onBlur() {
  blurTimer = setTimeout(() => { focused.value = false }, 150)
}
</script>

<template>
  <div class="flex-1 flex justify-center">
    <div class="relative w-full max-w-md">
      <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
      <input
        v-model="query"
        type="text"
        placeholder="Search pages..."
        class="w-full pl-9 pr-3 py-1.5 text-sm rounded-lg border border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-900 text-zinc-900 dark:text-zinc-100 placeholder-zinc-400 dark:placeholder-zinc-500 outline-none focus:border-zinc-400 dark:focus:border-zinc-600 focus:ring-1 focus:ring-zinc-400/20 dark:focus:ring-zinc-600/20 transition-colors"
        @focus="focused = true"
        @blur="onBlur"
        @keydown="onKeydown"
      />
      <div
        v-if="showDropdown"
        class="absolute top-full left-0 right-0 mt-1 rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-lg max-h-80 overflow-y-auto z-50"
      >
        <template v-if="results.length">
          <button
            v-for="(result, i) in results"
            :key="result.path"
            :data-search-selected="i === selectedIndex"
            class="flex flex-col w-full px-3 py-2 text-left transition-colors"
            :class="i === selectedIndex
              ? 'bg-zinc-100 dark:bg-zinc-800'
              : 'hover:bg-zinc-100 dark:hover:bg-zinc-800'"
            @mousedown.prevent="navigate(result.path)"
            @mouseenter="selectedIndex = i"
          >
            <span class="text-sm text-zinc-900 dark:text-zinc-100">{{ result.label }}</span>
            <span class="text-[11px] text-zinc-400 dark:text-zinc-500 truncate">{{ result.breadcrumb }}</span>
          </button>
        </template>
        <div v-else class="px-3 py-4 text-sm text-zinc-400 dark:text-zinc-500 text-center">
          No pages found
        </div>
      </div>
    </div>
  </div>
</template>
