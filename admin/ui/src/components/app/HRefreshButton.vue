<script setup lang="ts">
import { inject } from 'vue'
import { RefreshCw } from 'lucide-vue-next'

const props = defineProps<{ loading?: boolean }>()
const reload = inject<(() => void) | null>('hotpot-table-reload', null)
</script>

<template>
  <button
    v-if="reload"
    :disabled="loading"
    class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-zinc-700 dark:text-zinc-300 border border-zinc-300 dark:border-zinc-600 rounded-md hover:bg-zinc-50 dark:hover:bg-zinc-800 disabled:opacity-50 transition-colors shrink-0"
    @click="reload"
  >
    <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" />
    Refresh
  </button>
</template>
