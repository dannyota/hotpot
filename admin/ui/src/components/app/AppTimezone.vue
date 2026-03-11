<script setup lang="ts">
import { ref } from 'vue'
import { Globe } from 'lucide-vue-next'
import { useTimezone, TIMEZONE_OPTIONS } from '@/composables/useTimezone'

const { timezone, label, setTimezone } = useTimezone()
const open = ref(false)

function select(tz: string) {
  setTimezone(tz)
  open.value = false
}
</script>

<template>
  <div class="relative">
    <button
      class="inline-flex items-center gap-1 px-2 py-1.5 rounded-md text-xs font-medium text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
      @click.stop="open = !open"
    >
      <Globe class="w-3.5 h-3.5" />
      <span>{{ label }}</span>
    </button>

    <Teleport to="body">
      <div v-if="open" class="fixed inset-0 z-40" @click="open = false" />
    </Teleport>

    <div
      v-if="open"
      class="absolute right-0 top-full mt-1 z-50 w-44 rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-xl py-1 max-h-80 overflow-y-auto"
    >
      <button
        v-for="opt in TIMEZONE_OPTIONS"
        :key="opt.value"
        class="flex items-center w-full px-3 py-1.5 text-xs text-left transition-colors"
        :class="timezone === opt.value
          ? 'bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 font-medium'
          : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800'"
        @click="select(opt.value)"
      >
        {{ opt.label }}
      </button>
    </div>
  </div>
</template>
