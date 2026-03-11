<script setup lang="ts">
defineProps<{
  label: string
  value: string | number
  subtitle?: string
  color?: string
  icon?: any
}>()
</script>

<template>
  <div
    class="group relative flex items-start gap-4 p-4 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 hover:border-zinc-300 dark:hover:border-zinc-700 hover:shadow-sm transition-all text-left"
  >
    <div
      v-if="icon"
      class="w-10 h-10 rounded-lg flex items-center justify-center shrink-0"
      :style="{ backgroundColor: (color ?? '#888') + '14' }"
    >
      <component :is="icon" class="w-5 h-5" :style="{ color: color }" />
    </div>
    <div class="flex-1 min-w-0">
      <span class="text-sm font-medium text-zinc-900 dark:text-zinc-100">{{ label }}</span>
      <div class="flex items-baseline gap-3 mt-1">
        <span class="text-2xl font-semibold text-zinc-900 dark:text-zinc-100 tabular-nums">
          {{ typeof value === 'number' ? value.toLocaleString() : value }}
        </span>
        <span v-if="subtitle" class="text-xs text-zinc-400 dark:text-zinc-500">{{ subtitle }}</span>
      </div>
    </div>
    <slot />
  </div>
</template>
