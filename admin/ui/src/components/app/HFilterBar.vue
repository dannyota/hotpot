<script setup lang="ts">
import { ref } from 'vue'
import { Search, X } from 'lucide-vue-next'

const props = defineProps<{
  searchPlaceholder?: string
  hasActiveFilters?: boolean
}>()

const emit = defineEmits<{
  search: [value: string]
  clear: []
}>()

const searchValue = ref('')

function onSearch() {
  emit('search', searchValue.value)
}

function onClear() {
  searchValue.value = ''
  emit('clear')
}
</script>

<template>
  <div class="flex items-center gap-2 flex-wrap">
    <div class="relative flex-1 max-w-xs">
      <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
      <input
        v-model="searchValue"
        type="text"
        :placeholder="searchPlaceholder ?? 'Search...'"
        class="w-full pl-8 pr-3 py-1.5 text-sm border border-zinc-200 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-900 text-zinc-900 dark:text-zinc-100 placeholder:text-zinc-400 focus:outline-none focus:ring-2 focus:ring-zinc-900/10 dark:focus:ring-zinc-100/10 transition-shadow"
        @input="onSearch"
      />
    </div>
    <slot />
    <button
      v-if="hasActiveFilters"
      class="inline-flex items-center gap-1 px-2 py-1.5 text-xs text-zinc-500 hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-200 transition-colors"
      @click="onClear"
    >
      <X class="w-3 h-3" />Clear
    </button>
  </div>
</template>
