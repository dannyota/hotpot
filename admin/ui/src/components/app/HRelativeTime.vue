<script setup lang="ts">
import { TooltipContent, TooltipPortal, TooltipRoot, TooltipTrigger } from 'reka-ui'
import { useTimezone } from '@/composables/useTimezone'

defineProps<{ value: string | null | undefined }>()
const { formatDateTime, relativeTime } = useTimezone()
</script>

<template>
  <TooltipRoot :delay-duration="200">
    <TooltipTrigger as-child>
      <span class="text-zinc-500 dark:text-zinc-400 text-xs tabular-nums cursor-default">{{ value ? relativeTime(value) : '' }}</span>
    </TooltipTrigger>
    <TooltipPortal>
      <TooltipContent
        side="top"
        :side-offset="4"
        class="z-[100] rounded-md border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-lg px-2.5 py-1.5 text-xs text-zinc-700 dark:text-zinc-300 animate-in fade-in-0 zoom-in-95"
      >
        {{ value ? formatDateTime(value) : '' }}
      </TooltipContent>
    </TooltipPortal>
  </TooltipRoot>
</template>
