<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="flex items-start justify-between">
    <div>
      <h1 class="text-xl font-semibold text-zinc-900 dark:text-zinc-100">{{ title }}</h1>
      <p v-if="description" class="text-sm text-zinc-500 dark:text-zinc-400 mt-1">{{ description }}</p>
    </div>
    <div class="flex items-center gap-2">
      <slot name="actions" />
    </div>
  </div>
  <div class="h-px bg-zinc-200 dark:bg-zinc-800" />
</template>
