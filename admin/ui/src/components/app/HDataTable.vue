<script setup lang="ts" generic="T">
import { computed, ref } from 'vue'
import { ArrowUpDown, ArrowUp, ArrowDown, ChevronLeft, ChevronRight } from 'lucide-vue-next'
import type { PaginationMeta } from '@/types/api'
import HAlert from '@/components/app/HAlert.vue'

export interface Column<T> {
  key: string
  label: string
  sortable?: boolean
  maxWidth?: number
}

const props = defineProps<{
  columns: Column<T>[]
  data: T[]
  meta: PaginationMeta
  sort?: string
  loading?: boolean
  error?: string | null
  rowKey?: keyof T | ((row: T) => string)
  pageSizes?: number[]
}>()

const emit = defineEmits<{
  sort: [field: string]
  page: [page: number]
  'page-size': [size: number]
  'row-click': [row: T]
}>()

const sizes = computed(() => props.pageSizes ?? [10, 20, 50, 100])
const customSize = ref('')
const showCustomInput = ref(false)

function applyCustomSize() {
  const n = parseInt(customSize.value, 10)
  if (n > 0 && n <= 500) {
    emit('page-size', n)
    showCustomInput.value = false
    customSize.value = ''
  }
}

function onSizeChange(e: Event) {
  const val = (e.target as HTMLSelectElement).value
  if (val === 'custom') {
    showCustomInput.value = true
    return
  }
  emit('page-size', Number(val))
}

function getRowKey(row: T, index: number): string {
  if (!props.rowKey) return String(index)
  if (typeof props.rowKey === 'function') return props.rowKey(row)
  return String(row[props.rowKey])
}

function sortIcon(field: string) {
  if (props.sort === field) return ArrowUp
  if (props.sort === `-${field}`) return ArrowDown
  return ArrowUpDown
}

function sortIconClass(field: string) {
  if (props.sort === field || props.sort === `-${field}`) {
    return 'w-3.5 h-3.5 text-zinc-900 dark:text-zinc-100'
  }
  return 'w-3.5 h-3.5 text-zinc-300 dark:text-zinc-600'
}

const pages = computed(() => {
  const total = props.meta.total_pages
  const current = props.meta.page
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  // Show first, last, and pages around current.
  const s = new Set([1, total, current - 1, current, current + 1].filter(p => p >= 1 && p <= total))
  return [...s].sort((a, b) => a - b)
})

const showStart = computed(() => {
  return (props.meta.page - 1) * props.meta.size + 1
})

const showEnd = computed(() => {
  return Math.min(props.meta.page * props.meta.size, props.meta.total)
})
</script>

<template>
  <div class="rounded-xl border border-zinc-200 dark:border-zinc-800 overflow-hidden bg-white dark:bg-zinc-900">
    <div v-if="loading" class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-zinc-200 dark:border-zinc-800">
            <th
              v-for="col in columns"
              :key="col.key"
              class="text-left px-4 py-3 text-xs font-medium text-zinc-500 dark:text-zinc-400 tracking-wider whitespace-nowrap"
            >{{ col.label }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-100 dark:divide-zinc-800/50">
          <tr v-for="i in 8" :key="i">
            <td v-for="col in columns" :key="col.key" class="px-4 py-3">
              <div class="h-4 rounded bg-zinc-200 dark:bg-zinc-800 animate-pulse" :style="{ width: (40 + ((i * 7 + columns.indexOf(col) * 13) % 40)) + '%' }" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <template v-else>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-zinc-200 dark:border-zinc-800">
              <th
                v-for="col in columns"
                :key="col.key"
                class="text-left px-4 py-3 text-xs font-medium text-zinc-500 dark:text-zinc-400 tracking-wider whitespace-nowrap"
              >
                <button
                  v-if="col.sortable"
                  class="inline-flex items-center gap-1 hover:text-zinc-700 dark:hover:text-zinc-200 transition-colors"
                  @click="emit('sort', col.key)"
                >
                  {{ col.label }}
                  <component :is="sortIcon(col.key)" :class="sortIconClass(col.key)" />
                </button>
                <span v-else>{{ col.label }}</span>
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-100 dark:divide-zinc-800/50">
            <tr
              v-for="(row, idx) in data"
              :key="getRowKey(row, idx)"
              class="hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors cursor-pointer"
              @click="emit('row-click', row)"
            >
              <td
                v-for="col in columns"
                :key="col.key"
                class="px-4 py-3 whitespace-nowrap"
                :class="{ 'overflow-hidden text-ellipsis': col.maxWidth }"
                :style="col.maxWidth ? { maxWidth: col.maxWidth + 'px' } : undefined"
                :title="col.maxWidth ? String((row as any)[col.key] ?? '') : undefined"
              >
                <slot :name="col.key" :row="row" :value="(row as any)[col.key]">
                  {{ (row as any)[col.key] }}
                </slot>
              </td>
            </tr>
            <tr v-if="data.length === 0">
              <td :colspan="columns.length" class="px-4 py-8">
                <div v-if="error" class="max-w-lg mx-auto">
                  <HAlert type="error" :message="error" />
                </div>
                <div v-else class="text-center text-zinc-400">
                  No data found
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div
        v-if="meta.total > 0"
        class="flex items-center justify-between px-4 py-3 border-t border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/50"
      >
        <div class="flex items-center gap-3">
          <span class="text-xs text-zinc-500 dark:text-zinc-400">
            Showing
            <span class="font-medium text-zinc-700 dark:text-zinc-300">{{ showStart }}&ndash;{{ showEnd }}</span>
            of
            <span class="font-medium text-zinc-700 dark:text-zinc-300">{{ meta.total }}</span>
          </span>
          <select
            v-if="!showCustomInput"
            :value="sizes.includes(meta.size) ? meta.size : 'custom'"
            class="text-xs border border-zinc-200 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 px-2 py-1 outline-none focus:ring-1 focus:ring-zinc-400/20"
            @change="onSizeChange"
          >
            <option v-for="s in sizes" :key="s" :value="s">{{ s }} / page</option>
            <option v-if="!sizes.includes(meta.size)" :value="meta.size">{{ meta.size }} / page</option>
            <option value="custom">Custom...</option>
          </select>
          <form v-else class="inline-flex items-center gap-1" @submit.prevent="applyCustomSize">
            <input
              v-model="customSize"
              type="number"
              min="1"
              max="500"
              placeholder="1-500"
              class="w-16 text-xs border border-zinc-200 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 px-2 py-1 outline-none focus:ring-1 focus:ring-zinc-400/20"
              autofocus
              @keydown.escape="showCustomInput = false"
            />
            <button type="submit" class="text-xs text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300">OK</button>
          </form>
        </div>
        <div class="flex items-center gap-1">
          <button
            :disabled="meta.page <= 1"
            class="p-1.5 rounded-md border border-zinc-200 dark:border-zinc-700 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            @click="emit('page', meta.page - 1)"
          >
            <ChevronLeft class="w-4 h-4" />
          </button>
          <template v-for="(n, i) in pages" :key="n">
            <span
              v-if="i > 0 && n - pages[i - 1] > 1"
              class="w-8 h-8 flex items-center justify-center text-xs text-zinc-400"
            >...</span>
            <button
              class="w-8 h-8 rounded-md text-xs font-medium transition-colors"
              :class="n === meta.page
                ? 'bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900'
                : 'text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800'"
              @click="emit('page', n)"
            >
              {{ n }}
            </button>
          </template>
          <button
            :disabled="meta.page >= meta.total_pages"
            class="p-1.5 rounded-md border border-zinc-200 dark:border-zinc-700 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
            @click="emit('page', meta.page + 1)"
          >
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </template>
  </div>
</template>
