<script setup lang="ts">
import { computed } from 'vue'
import { X, Copy, Check } from 'lucide-vue-next'
import { ref } from 'vue'

const props = defineProps<{
  data: any
  title?: string
}>()

const emit = defineEmits<{ close: [] }>()

const formatted = computed(() => {
  try {
    return JSON.stringify(props.data, null, 2)
  } catch {
    return String(props.data)
  }
})

const copied = ref(false)

async function copyToClipboard() {
  try {
    await navigator.clipboard.writeText(formatted.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1500)
  } catch { /* ignore */ }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="emit('close')" />
      <div class="relative bg-white dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-700 shadow-2xl w-full max-w-2xl max-h-[80vh] flex flex-col mx-4">
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-zinc-200 dark:border-zinc-800">
          <span class="text-sm font-medium text-zinc-900 dark:text-zinc-100">{{ title ?? 'JSON Detail' }}</span>
          <div class="flex items-center gap-1">
            <button
              class="p-1.5 rounded-md text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
              @click="copyToClipboard"
            >
              <Check v-if="copied" class="w-4 h-4 text-green-500" />
              <Copy v-else class="w-4 h-4" />
            </button>
            <button
              class="p-1.5 rounded-md text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
              @click="emit('close')"
            >
              <X class="w-4 h-4" />
            </button>
          </div>
        </div>
        <!-- Body -->
        <div class="flex-1 overflow-auto p-4">
          <pre class="text-xs text-zinc-700 dark:text-zinc-300 whitespace-pre-wrap break-words font-mono leading-relaxed">{{ formatted }}</pre>
        </div>
      </div>
    </div>
  </Teleport>
</template>
