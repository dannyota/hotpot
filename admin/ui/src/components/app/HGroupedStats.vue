<script setup lang="ts">
defineProps<{
  stats: Record<string, { count: number; breakdown: Record<string, number> }>
  total: number
  dotColors: Record<string, string>
  defaultDotColor?: string
  uppercaseBreakdown?: boolean
}>()
</script>

<template>
  <div class="flex items-center gap-2 text-sm flex-wrap">
    <template v-for="(group, status, i) in stats" :key="status">
      <span v-if="i > 0" class="text-zinc-300 dark:text-zinc-600">&bull;</span>
      <span class="inline-flex items-center gap-1">
        <span
          class="w-1.5 h-1.5 rounded-full"
          :class="dotColors[String(status).toUpperCase()] ?? defaultDotColor ?? 'bg-zinc-300 dark:bg-zinc-600'"
        />
        <span class="font-semibold text-zinc-900 dark:text-zinc-100">{{ group.count }}</span>
        <span class="text-zinc-600 dark:text-zinc-400">{{ status }}</span>
        <span class="text-zinc-400 dark:text-zinc-500">({{
          Object.entries(group.breakdown)
            .sort(([, a], [, b]) => b - a)
            .map(([k, v]) => `${uppercaseBreakdown ? k.toUpperCase() : k}: ${v}`)
            .join(', ')
        }})</span>
      </span>
    </template>
    <span class="text-zinc-300 dark:text-zinc-600">&bull;</span>
    <span class="text-zinc-500 dark:text-zinc-400">
      TOTAL: <span class="font-semibold text-zinc-900 dark:text-zinc-100">{{ total.toLocaleString() }}</span>
    </span>
  </div>
</template>
