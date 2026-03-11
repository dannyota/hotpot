<script setup lang="ts">
import { computed } from 'vue'
import { AlertCircle, AlertTriangle, Info, X } from 'lucide-vue-next'

const props = defineProps<{
  type?: 'error' | 'warning' | 'info'
  message: string
  dismissible?: boolean
}>()

const emit = defineEmits<{ dismiss: [] }>()

const config = {
  error: {
    icon: AlertCircle,
    bg: 'bg-red-50 dark:bg-red-950/30',
    border: 'border-red-200 dark:border-red-900/50',
    text: 'text-red-800 dark:text-red-300',
    iconColor: 'text-red-500 dark:text-red-400',
  },
  warning: {
    icon: AlertTriangle,
    bg: 'bg-amber-50 dark:bg-amber-950/30',
    border: 'border-amber-200 dark:border-amber-900/50',
    text: 'text-amber-800 dark:text-amber-300',
    iconColor: 'text-amber-500 dark:text-amber-400',
  },
  info: {
    icon: Info,
    bg: 'bg-blue-50 dark:bg-blue-950/30',
    border: 'border-blue-200 dark:border-blue-900/50',
    text: 'text-blue-800 dark:text-blue-300',
    iconColor: 'text-blue-500 dark:text-blue-400',
  },
}

const c = computed(() => config[props.type ?? 'error'])
</script>

<template>
  <div
    class="flex items-start gap-3 px-4 py-3 rounded-lg border text-sm"
    :class="[c.bg, c.border, c.text]"
  >
    <component :is="c.icon" class="w-4 h-4 mt-0.5 shrink-0" :class="c.iconColor" />
    <span class="flex-1">{{ message }}</span>
    <button
      v-if="dismissible"
      class="shrink-0 p-0.5 rounded hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
      @click="emit('dismiss')"
    >
      <X class="w-3.5 h-3.5" />
    </button>
  </div>
</template>
