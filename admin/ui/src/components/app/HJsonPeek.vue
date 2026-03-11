<script setup lang="ts">
import { computed } from 'vue'
import {
  TooltipRoot,
  TooltipTrigger,
  TooltipPortal,
  TooltipContent,
} from 'reka-ui'

const props = defineProps<{
  /** Label shown before the {...} button (e.g. "Redhat", "172.19.1.8"). */
  label?: string
  /** Extra count badge (e.g. +3). */
  extra?: number
  /** Data to preview on hover and show on click. */
  data: any
  /** Title for the full JSON viewer modal. */
  title?: string
}>()

const emit = defineEmits<{ click: [data: any, title: string] }>()

const formatted = computed(() => {
  try {
    return JSON.stringify(props.data, null, 2)
  } catch {
    return String(props.data)
  }
})

const preview = computed(() => {
  const text = formatted.value
  if (text.length <= 500) return text
  return text.slice(0, 500) + '\n...'
})
</script>

<template>
  <span v-if="data" class="inline-flex items-center gap-1">
    <span v-if="label" class="text-sm">{{ label }}</span>
    <span
      v-if="extra && extra > 0"
      class="text-[10px] px-1 py-0.5 rounded-full bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400 font-medium"
    >+{{ extra }}</span>
    <TooltipRoot :delay-duration="400" :disable-closing-trigger="false">
      <TooltipTrigger as-child>
        <button
          class="text-xs px-1 py-0.5 rounded border border-zinc-200 dark:border-zinc-700 text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors font-mono"
          @click.stop="emit('click', data, title ?? '')"
        >{...}</button>
      </TooltipTrigger>
      <TooltipPortal>
        <TooltipContent
          side="bottom"
          :side-offset="4"
          align="start"
          class="z-[100] max-w-md max-h-72 overflow-auto rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-xl p-3 animate-in fade-in-0 zoom-in-95"
        >
          <pre class="text-xs text-zinc-700 dark:text-zinc-300 whitespace-pre-wrap break-words font-mono leading-relaxed">{{ preview }}</pre>
        </TooltipContent>
      </TooltipPortal>
    </TooltipRoot>
  </span>
  <span v-else class="text-zinc-400">—</span>
</template>
