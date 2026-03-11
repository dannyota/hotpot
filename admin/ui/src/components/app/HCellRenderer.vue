<script setup lang="ts">
import type { ColumnDef } from '@/types/table'
import { columnLabel } from '@/composables/columns'
import { displayValue, boolBadgeClass } from '@/composables/formatting'
import HDateTime from '@/components/app/HDateTime.vue'
import HRelativeTime from '@/components/app/HRelativeTime.vue'
import HJsonPeek from '@/components/app/HJsonPeek.vue'

defineProps<{
  col: ColumnDef
  value: any
  row: any
}>()

const emit = defineEmits<{
  'show-json': [data: any, title: string]
}>()
</script>

<template>
  <HDateTime v-if="col.format === 'date'" :value="value" />
  <HRelativeTime v-else-if="col.format === 'relative'" :value="value" />
  <span v-else-if="col.format === 'bool'" class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full" :class="boolBadgeClass(value)">{{ value ? 'Yes' : 'No' }}</span>
  <HJsonPeek v-else-if="col.format === 'json'" :data="value" :title="col.label ?? columnLabel(col.key)" @click="(d: any, t: string) => emit('show-json', d, t)" />
  <span v-else-if="col.badge" class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full" :class="col.badge(value)">{{ displayValue(col, value, row) }}</span>
  <span v-else-if="col.format === 'bold'" class="font-medium text-zinc-900 dark:text-zinc-100">{{ displayValue(col, value, row) }}</span>
  <span v-else-if="col.format === 'mono'" class="font-mono text-xs text-zinc-500 dark:text-zinc-400">{{ displayValue(col, value, row) }}</span>
  <span v-else-if="col.format === 'number'" class="tabular-nums text-sm">{{ displayValue(col, value, row) }}</span>
  <template v-else>{{ displayValue(col, value, row) }}</template>
</template>
