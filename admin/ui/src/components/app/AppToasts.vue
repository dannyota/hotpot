<script setup lang="ts">
import { AlertCircle, AlertTriangle, Info, X } from 'lucide-vue-next'
import { useNotifications, shortenSource } from '@/composables/useNotifications'

const { toasts, dismissToast } = useNotifications()

const typeConfig = {
  error: { icon: AlertCircle, bg: 'bg-red-600 dark:bg-red-700', iconColor: 'text-red-200' },
  warning: { icon: AlertTriangle, bg: 'bg-amber-600 dark:bg-amber-700', iconColor: 'text-amber-200' },
  info: { icon: Info, bg: 'bg-blue-600 dark:bg-blue-700', iconColor: 'text-blue-200' },
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 w-80 pointer-events-none">
      <TransitionGroup
        enter-active-class="transition-all duration-300 ease-out"
        leave-active-class="transition-all duration-200 ease-in"
        enter-from-class="translate-x-full opacity-0"
        enter-to-class="translate-x-0 opacity-100"
        leave-from-class="translate-x-0 opacity-100"
        leave-to-class="translate-x-full opacity-0"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto rounded-lg shadow-lg text-white px-4 py-3 flex items-start gap-3"
          :class="typeConfig[toast.type].bg"
        >
          <component
            :is="typeConfig[toast.type].icon"
            class="w-4 h-4 mt-0.5 shrink-0"
            :class="typeConfig[toast.type].iconColor"
          />
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium leading-snug">{{ toast.message }}</p>
            <p class="text-[11px] text-white/60 mt-0.5 truncate">{{ shortenSource(toast.source) }}</p>
          </div>
          <button
            class="shrink-0 p-0.5 rounded hover:bg-white/10 transition-colors"
            @click="dismissToast(toast.id)"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
