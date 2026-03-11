<script setup lang="ts">
import { useTimezone } from '@/composables/useTimezone'

defineProps<{ value: string | null | undefined }>()
const { formatDateTime } = useTimezone()
</script>

<template>
  <span class="text-zinc-500 dark:text-zinc-400 text-xs tabular-nums">{{ value ? formatDateTime(value) : '' }}</span>
</template>
