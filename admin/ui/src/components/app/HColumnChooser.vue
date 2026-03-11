<script setup lang="ts">
import { ref, computed } from 'vue'
import { Columns3 } from 'lucide-vue-next'
import type { Column } from '@/components/app/HDataTable.vue'

const props = defineProps<{
  columns: Column<any>[]
  hiddenKeys: Set<string>
}>()

const emit = defineEmits<{
  toggle: [key: string, visible: boolean]
  reset: []
}>()

const open = ref(false)

function onBlur() {
  setTimeout(() => { open.value = false }, 150)
}

const hasHidden = computed(() => props.hiddenKeys.size > 0)
</script>

<template>
  <div class="relative" @focusout="onBlur">
    <button
      class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium border rounded-md transition-colors shrink-0"
      :class="hasHidden
        ? 'border-zinc-400 dark:border-zinc-500 text-zinc-900 dark:text-zinc-100 bg-white dark:bg-zinc-900'
        : 'border-zinc-200 dark:border-zinc-700 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800'"
      @click="open = !open"
    >
      <Columns3 class="w-3.5 h-3.5" />
      Columns
      <span
        v-if="hasHidden"
        class="ml-0.5 text-[10px] tabular-nums text-zinc-500 dark:text-zinc-400"
      >{{ columns.length - hiddenKeys.size }}/{{ columns.length }}</span>
    </button>

    <div
      v-if="open"
      class="absolute top-full right-0 mt-1 w-56 max-h-72 overflow-y-auto rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-lg z-40"
    >
      <!-- Reset -->
      <div
        v-if="hasHidden"
        class="flex items-center justify-end px-3 py-1.5 border-b border-zinc-100 dark:border-zinc-800"
      >
        <button
          class="text-[11px] text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300"
          @mousedown.prevent="emit('reset')"
        >
          Reset to default
        </button>
      </div>
      <!-- Column list -->
      <label
        v-for="col in columns"
        :key="col.key"
        class="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors"
        @mousedown.prevent="emit('toggle', col.key, hiddenKeys.has(col.key))"
      >
        <input
          type="checkbox"
          :checked="!hiddenKeys.has(col.key)"
          class="rounded border-zinc-300 dark:border-zinc-600 text-zinc-900 dark:text-zinc-100 w-3.5 h-3.5 pointer-events-none"
        />
        <span class="truncate text-zinc-700 dark:text-zinc-300 text-xs">{{ col.label }}</span>
      </label>
    </div>
  </div>
</template>
