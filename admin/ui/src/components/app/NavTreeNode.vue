<script setup lang="ts">
import { computed, inject, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ChevronRight, Star } from 'lucide-vue-next'
import type { NavItem } from '@/composables/useUIConfig'
import { useFavorites } from '@/composables/useFavorites'

const props = defineProps<{
  item: NavItem
  depth: number
  parentKey?: string
}>()

const route = useRoute()
const router = useRouter()
const expanded = inject<Ref<Set<string>>>('nav-expanded')!

const nodeKey = computed(() =>
  props.parentKey ? `${props.parentKey}/${props.item.label}` : props.item.label
)

const isExpanded = computed(() => expanded.value.has(nodeKey.value))

function toggle() {
  const s = new Set(expanded.value)
  if (s.has(nodeKey.value)) s.delete(nodeKey.value)
  else s.add(nodeKey.value)
  expanded.value = s
}

function leafCount(item: NavItem): number {
  if (!item.children) return 1
  return item.children.reduce((n, c) => n + leafCount(c), 0)
}

const { isFavorite, toggleFavorite } = useFavorites()

const isLeaf = computed(() => !props.item.children)
const isActive = computed(() => props.item.path === route.path)
const starred = computed(() => props.item.path ? isFavorite(props.item.path) : false)

function navigate() {
  if (props.item.path) router.push(props.item.path)
}

function onToggleFavorite(e: Event) {
  e.stopPropagation()
  if (props.item.path) toggleFavorite(props.item.path)
}

const indent = computed(() => `${props.depth * 12 + 16}px`)
</script>

<template>
  <!-- Leaf -->
  <button
    v-if="isLeaf"
    class="group flex items-center w-full rounded-md text-sm transition-colors pr-1.5 py-1.5"
    :style="{ paddingLeft: indent }"
    :class="isActive
      ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
      : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:text-zinc-100 dark:hover:bg-zinc-800'"
    @click="navigate"
  >
    <span class="truncate flex-1 text-left">{{ item.label }}</span>
    <span
      class="shrink-0 p-1 rounded transition-colors"
      :class="starred
        ? 'text-amber-400 opacity-100'
        : 'opacity-0 group-hover:opacity-100 text-zinc-400 hover:text-amber-400'"
      @click="onToggleFavorite"
    >
      <Star class="w-3 h-3" :fill="starred ? 'currentColor' : 'none'" />
    </span>
  </button>

  <!-- Group -->
  <div v-else>
    <button
      class="group flex items-center gap-1.5 w-full rounded-md text-[13px] transition-colors pr-1.5 py-1.5 text-zinc-500 hover:text-zinc-900 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:text-zinc-100 dark:hover:bg-zinc-800"
      :style="{ paddingLeft: indent }"
      @click="toggle"
    >
      <ChevronRight
        class="w-3 h-3 transition-transform shrink-0"
        :class="isExpanded && 'rotate-90'"
      />
      <span class="truncate flex-1 text-left font-medium">{{ item.label }}</span>
      <span class="text-[11px] text-zinc-400 dark:text-zinc-500 tabular-nums shrink-0">{{ leafCount(item) }}</span>
      <span
        class="shrink-0 p-0.5 rounded transition-colors"
        :class="isFavorite(nodeKey)
          ? 'text-amber-400 opacity-100'
          : 'opacity-0 group-hover:opacity-100 text-zinc-400 hover:text-amber-400'"
        @click.stop="toggleFavorite(nodeKey)"
      >
        <Star class="w-3 h-3" :fill="isFavorite(nodeKey) ? 'currentColor' : 'none'" />
      </span>
    </button>
    <div v-if="isExpanded" class="space-y-0.5">
      <NavTreeNode
        v-for="child in item.children"
        :key="child.label"
        :item="child"
        :depth="depth + 1"
        :parent-key="nodeKey"
      />
    </div>
  </div>
</template>
