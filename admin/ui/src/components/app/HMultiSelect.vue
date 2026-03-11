<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { ChevronDown, X } from 'lucide-vue-next'
import type { FilterOption } from '@/types/api'

const props = defineProps<{
  label: string
  options: FilterOption[]
  modelValue: string[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const container = ref<HTMLElement | null>(null)
const open = ref(false)

// Local draft that tracks selections while the dropdown is open.
const draft = ref<string[]>([...props.modelValue])

// Sync draft when modelValue changes externally (e.g. clear all, restored filters).
watch(() => props.modelValue, (v) => {
  draft.value = [...v]
})

function isDrafted(val: string): boolean {
  return draft.value.includes(val)
}

function apply() {
  const next = draft.value
  if (
    next.length !== props.modelValue.length ||
    !next.every(v => props.modelValue.includes(v))
  ) {
    emit('update:modelValue', [...next])
  }
}

function closeAndApply() {
  if (!open.value) return
  open.value = false
  apply()
}

// Click-outside detection: close and apply when clicking anywhere outside the component.
function onDocumentClick(e: MouseEvent) {
  if (container.value && !container.value.contains(e.target as Node)) {
    closeAndApply()
  }
}

watch(open, (isOpen) => {
  if (isOpen) {
    draft.value = [...props.modelValue]
    document.addEventListener('mousedown', onDocumentClick)
  } else {
    document.removeEventListener('mousedown', onDocumentClick)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocumentClick)
})

function toggleOpen() {
  if (open.value) {
    closeAndApply()
  } else {
    open.value = true
  }
}

function toggle(val: string) {
  if (draft.value.includes(val)) {
    draft.value = draft.value.filter(v => v !== val)
  } else {
    draft.value = [...draft.value, val]
  }
}

function clear() {
  draft.value = []
  apply()
}

function selectAll() {
  draft.value = props.options.map(o => o.value)
}

const displayLabel = computed(() => {
  const len = draft.value.length
  if (len === 0) return props.label
  if (len === 1) {
    const opt = props.options.find(o => o.value === draft.value[0])
    return opt?.label ?? draft.value[0]
  }
  return `${len} selected`
})
</script>

<template>
  <div ref="container" class="relative">
    <button
      class="inline-flex items-center gap-1.5 px-2.5 py-1.5 text-sm border rounded-md transition-colors min-w-[7rem]"
      :class="draft.length > 0
        ? 'border-zinc-400 dark:border-zinc-500 text-zinc-900 dark:text-zinc-100 bg-white dark:bg-zinc-900'
        : 'border-zinc-200 dark:border-zinc-700 text-zinc-500 dark:text-zinc-400 bg-white dark:bg-zinc-900'"
      @click="toggleOpen"
    >
      <span class="truncate flex-1 text-left text-xs">{{ displayLabel }}</span>
      <X v-if="draft.length > 0" class="w-3 h-3 shrink-0 hover:text-red-500" @click.stop="clear" />
      <ChevronDown v-else class="w-3 h-3 shrink-0" />
    </button>

    <div
      v-if="open"
      class="absolute top-full left-0 mt-1 w-52 max-h-60 overflow-y-auto rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-lg z-40"
    >
      <!-- Select all / Clear -->
      <div class="flex items-center justify-between px-3 py-1.5 border-b border-zinc-100 dark:border-zinc-800">
        <button class="text-[11px] text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300" @mousedown.prevent="selectAll">
          Select all
        </button>
        <button class="text-[11px] text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300" @mousedown.prevent="clear">
          Clear
        </button>
      </div>
      <!-- Options -->
      <label
        v-for="opt in options"
        :key="opt.value"
        class="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors"
        @click.prevent="toggle(opt.value)"
      >
        <input
          type="checkbox"
          :checked="isDrafted(opt.value)"
          class="rounded border-zinc-300 dark:border-zinc-600 text-zinc-900 dark:text-zinc-100 w-3.5 h-3.5 pointer-events-none"
        />
        <span class="truncate text-zinc-700 dark:text-zinc-300 text-xs flex-1">{{ opt.label ?? opt.value }}</span>
        <span class="text-[11px] text-zinc-400 dark:text-zinc-500 tabular-nums">{{ opt.count }}</span>
      </label>
      <div v-if="options.length === 0" class="px-3 py-3 text-xs text-zinc-400 text-center">
        No options
      </div>
    </div>
  </div>
</template>
