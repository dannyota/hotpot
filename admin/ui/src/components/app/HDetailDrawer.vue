<script setup lang="ts">
import { X, Copy, Check } from 'lucide-vue-next'
import { ref } from 'vue'

defineProps<{
  title?: string
  fields: { label: string; value: any; mono?: boolean }[]
}>()

const emit = defineEmits<{ close: [] }>()

const copiedField = ref('')

async function copyValue(label: string, value: any) {
  try {
    const text = typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)
    await navigator.clipboard.writeText(text)
    copiedField.value = label
    setTimeout(() => { copiedField.value = '' }, 1500)
  } catch { /* ignore */ }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex justify-end">
      <div class="absolute inset-0 bg-black/30" @click="emit('close')" />
      <div class="relative bg-white dark:bg-zinc-900 border-l border-zinc-200 dark:border-zinc-700 shadow-2xl w-full max-w-lg 2xl:max-w-xl flex flex-col animate-in slide-in-from-right">
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-zinc-200 dark:border-zinc-800 shrink-0">
          <span class="text-sm font-medium text-zinc-900 dark:text-zinc-100 truncate">{{ title ?? 'Details' }}</span>
          <button
            class="p-1.5 rounded-md text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
            @click="emit('close')"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
        <!-- Body -->
        <div class="flex-1 overflow-auto">
          <dl class="divide-y divide-zinc-100 dark:divide-zinc-800/50">
            <div
              v-for="field in fields"
              :key="field.label"
              class="px-4 py-3 flex justify-between gap-4 group"
            >
              <dt class="text-xs text-zinc-500 dark:text-zinc-400 shrink-0 pt-0.5">{{ field.label }}</dt>
              <dd class="text-sm text-zinc-900 dark:text-zinc-100 text-right min-w-0">
                <span class="inline-flex items-center gap-1">
                  <span :class="{ 'font-mono text-xs': field.mono }" class="break-all">
                    <template v-if="typeof field.value === 'object' && field.value !== null">
                      <pre class="text-xs whitespace-pre-wrap break-words font-mono leading-relaxed text-left">{{ JSON.stringify(field.value, null, 2) }}</pre>
                    </template>
                    <template v-else>{{ field.value ?? '—' }}</template>
                  </span>
                  <button
                    v-if="field.value != null && field.value !== ''"
                    class="p-0.5 rounded text-zinc-300 dark:text-zinc-600 hover:text-zinc-500 dark:hover:text-zinc-400 opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
                    @click="copyValue(field.label, field.value)"
                  >
                    <Check v-if="copiedField === field.label" class="w-3 h-3 text-green-500" />
                    <Copy v-else class="w-3 h-3" />
                  </button>
                </span>
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  </Teleport>
</template>
